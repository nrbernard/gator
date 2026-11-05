package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type PostService struct {
	Repo *database.Queries
}

// Queue returns the newest Post in the user's Queue and the Queue's size or,
// if the Queue is empty, when gator last fetched the user's Feeds.
func (s *PostService) Queue(ctx context.Context, userID uuid.UUID) (models.Queue, error) {
	count, err := s.Repo.CountQueuePosts(ctx, userID.String())
	if err != nil {
		return models.Queue{}, fmt.Errorf("count queue: %w", err)
	}
	queue := models.Queue{Count: int(count)}

	row, err := s.Repo.GetNextQueuePost(ctx, userID.String())
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

// CountQueueCreatedBy returns how many Queue Posts gator fetched at or before
// cutoff: the Posts PostStateService.TrashAll would trash.
func (s *PostService) CountQueueCreatedBy(ctx context.Context, userID uuid.UUID, cutoff time.Time) (int, error) {
	n, err := s.Repo.CountQueuePostsCreatedBy(ctx, database.CountQueuePostsCreatedByParams{
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
	rows, err := s.Repo.GetReadingListPosts(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("get reading list: %w", err)
	}
	posts := make([]models.Post, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, models.Post{
			ID:          uuid.MustParse(row.ID),
			Title:       row.Title,
			Link:        row.Url,
			Description: row.Description.String,
			PublishedAt: row.PublishedAt,
			FeedID:      uuid.MustParse(row.FeedID),
			FeedName:    row.FeedName,
			IsRead:      row.ReadAt.Valid,
		})
	}
	return posts, nil
}

// Archive returns the user's archived Posts, most recently archived first.
func (s *PostService) Archive(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	rows, err := s.Repo.GetArchivePosts(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("get archive: %w", err)
	}
	posts := make([]models.Post, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, models.Post{
			ID:          uuid.MustParse(row.ID),
			Title:       row.Title,
			Link:        row.Url,
			Description: row.Description.String,
			PublishedAt: row.PublishedAt,
			FeedID:      uuid.MustParse(row.FeedID),
			FeedName:    row.FeedName,
			IsRead:      row.ReadAt.Valid,
		})
	}
	return posts, nil
}

// Trash returns the user's trashed Posts, most recently trashed first.
func (s *PostService) Trash(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	rows, err := s.Repo.GetTrashPosts(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("get trash: %w", err)
	}
	posts := make([]models.Post, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, models.Post{
			ID:          uuid.MustParse(row.ID),
			Title:       row.Title,
			Link:        row.Url,
			Description: row.Description.String,
			PublishedAt: row.PublishedAt,
			FeedID:      uuid.MustParse(row.FeedID),
			FeedName:    row.FeedName,
			IsRead:      row.ReadAt.Valid,
		})
	}
	return posts, nil
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
		results = append(results, SearchResult{
			Post: models.Post{
				ID:          uuid.MustParse(row.ID),
				Title:       row.Title,
				Link:        row.Url,
				Description: row.Description.String,
				PublishedAt: row.PublishedAt,
				FeedID:      uuid.MustParse(row.FeedID),
				FeedName:    row.FeedName,
				IsRead:      row.ReadAt.Valid,
			},
			Place: placeOf(row.Place),
		})
	}
	return results, nil
}

// placeOf is the Place a post_states place column records: NULL is the Queue.
func placeOf(place sql.NullString) Place {
	if !place.Valid {
		return PlaceQueue
	}
	return Place(place.String)
}
