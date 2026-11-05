package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/clock"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type PostService struct {
	Repo  *database.Queries
	Clock *clock.Clock
}

// Cutoff is the cutoff for a Queue page loading now: the end of the last
// complete second. gator records a Post's fetch to the second, so a Post
// fetched in the current second, still open, counts as later.
func (s *PostService) Cutoff() time.Time {
	return s.Clock.Now().Truncate(time.Second).Add(-time.Second)
}

// AwaitSecond waits for the current second to end, so the next Cutoff covers
// every Post fetched so far.
func (s *PostService) AwaitSecond(ctx context.Context) error {
	return s.Clock.SleepUntil(ctx, s.Clock.Now().Truncate(time.Second).Add(time.Second))
}

// Queue returns the Queue as a page whose cutoff is cutoff sees it: the newest
// Post fetched by the cutoff, how many such Posts there are, and how many
// Posts are new since. If no Post is shown, it also returns when gator last
// fetched the user's Feeds.
func (s *PostService) Queue(ctx context.Context, userID uuid.UUID, cutoff time.Time) (models.Queue, error) {
	count, err := s.CountQueueAtOrBefore(ctx, userID, cutoff)
	if err != nil {
		return models.Queue{}, err
	}
	// The chip counts, and reveals, the Posts fetched since the cutoff up to
	// the cutoff a page loading now would have.
	newCutoff := s.Cutoff()
	newCount, err := s.Repo.CountQueuePostsCreatedBetween(ctx, database.CountQueuePostsCreatedBetweenParams{
		UserID:    userID.String(),
		Cutoff:    cutoff.UTC(),
		NewCutoff: newCutoff.UTC(),
	})
	if err != nil {
		return models.Queue{}, fmt.Errorf("count new posts: %w", err)
	}
	queue := models.Queue{Count: count, Cutoff: cutoff, NewCount: int(newCount), NewCutoff: newCutoff}

	row, err := s.Repo.GetNextQueuePost(ctx, database.GetNextQueuePostParams{
		UserID: userID.String(),
		Cutoff: cutoff.UTC(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		fetched, err := s.Repo.GetLastFetchedAt(ctx, userID.String())
		switch {
		case errors.Is(err, sql.ErrNoRows):
		case err != nil:
			return models.Queue{}, fmt.Errorf("get last fetch: %w", err)
		case fetched.Valid:
			queue.LastFetchedAt = &fetched.Time
		}
		return queue, nil
	}
	if err != nil {
		return models.Queue{}, fmt.Errorf("get next queue post: %w", err)
	}
	queue.Next = &models.Post{
		ID:          uuid.MustParse(row.ID),
		Title:       row.Title,
		Link:        row.Url,
		Description: row.Description.String,
		PublishedAt: row.PublishedAt,
		FeedID:      uuid.MustParse(row.FeedID),
		FeedName:    row.FeedName,
	}
	return queue, nil
}

// CountQueueAtOrBefore returns how many Queue Posts gator fetched at or before
// cutoff: the Posts PostStateService.TrashAll would trash.
func (s *PostService) CountQueueAtOrBefore(ctx context.Context, userID uuid.UUID, cutoff time.Time) (int, error) {
	n, err := s.Repo.CountQueuePostsCreatedAtOrBefore(ctx, database.CountQueuePostsCreatedAtOrBeforeParams{
		UserID: userID.String(),
		Cutoff: cutoff.UTC(),
	})
	if err != nil {
		return 0, fmt.Errorf("count queue: %w", err)
	}
	return int(n), nil
}

// ReadingList returns the user's kept Posts, unread first, then most recently
// kept first.
func (s *PostService) ReadingList(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	return s.postsIn(ctx, userID, PlaceReadingList)
}

// Archive returns the user's archived Posts, most recently archived first.
func (s *PostService) Archive(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	return s.postsIn(ctx, userID, PlaceArchive)
}

// Trash returns the user's trashed Posts, most recently trashed first.
func (s *PostService) Trash(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	return s.postsIn(ctx, userID, PlaceTrash)
}

// postsIn lists the Posts in a Place other than the Queue.
func (s *PostService) postsIn(ctx context.Context, userID uuid.UUID, place Place) ([]models.Post, error) {
	rows, err := s.Repo.GetPostsInPlace(ctx, database.GetPostsInPlaceParams{
		UserID: userID.String(),
		Place:  string(place),
	})
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", place.Label(), err)
	}
	posts := make([]models.Post, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, postFromRow(row))
	}
	return posts, nil
}

func postFromRow(row database.GetPostsInPlaceRow) models.Post {
	return models.Post{
		ID:          uuid.MustParse(row.ID),
		Title:       row.Title,
		Link:        row.Url,
		Description: row.Description.String,
		PublishedAt: row.PublishedAt,
		FeedID:      uuid.MustParse(row.FeedID),
		FeedName:    row.FeedName,
		IsRead:      row.ReadAt.Valid,
		Restorable:  Restorable(Place(row.TrashedFrom.String), row.FeedRemovedAt.Valid),
	}
}

// SearchResult is a Post found by search, with the Place it is in.
type SearchResult struct {
	models.Post
	Place Place
}

// SearchPosts returns the newest 100 Posts, in any Place, whose title or
// description contains query.
func (s *PostService) SearchPosts(ctx context.Context, userID uuid.UUID, query string) ([]SearchResult, error) {
	rows, err := s.Repo.SearchPostsByUser(ctx, database.SearchPostsByUserParams{
		UserID:     userID.String(),
		SearchText: query,
		LimitCount: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}

	results := make([]SearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, searchResultFromRow(row))
	}
	return results, nil
}

// SearchResult returns one Post as search shows it, with its current Place.
func (s *PostService) SearchResult(ctx context.Context, userID, postID uuid.UUID) (SearchResult, error) {
	row, err := s.Repo.GetSearchResult(ctx, database.GetSearchResultParams{
		UserID: userID.String(),
		PostID: postID.String(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return SearchResult{}, fmt.Errorf("%w: %s", ErrPostNotFound, postID)
	}
	if err != nil {
		return SearchResult{}, fmt.Errorf("get search result: %w", err)
	}
	// GetSearchResult returns the same columns as SearchPostsByUser.
	return searchResultFromRow(database.SearchPostsByUserRow(row)), nil
}

func searchResultFromRow(row database.SearchPostsByUserRow) SearchResult {
	return SearchResult{
		Post: models.Post{
			ID:          uuid.MustParse(row.ID),
			Title:       row.Title,
			Link:        row.Url,
			Description: row.Description.String,
			PublishedAt: row.PublishedAt,
			FeedID:      uuid.MustParse(row.FeedID),
			FeedName:    row.FeedName,
			IsRead:      row.ReadAt.Valid,
			Restorable:  Restorable(Place(row.TrashedFrom.String), row.FeedRemovedAt.Valid),
		},
		Place: placeOf(row.Place),
	}
}

// placeOf is the Place a post_states place column records: NULL is the Queue.
func placeOf(place sql.NullString) Place {
	if !place.Valid {
		return PlaceQueue
	}
	return Place(place.String)
}
