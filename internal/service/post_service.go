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

// Queue returns the Queue as a page whose cutoff is cutoff sees it: the newest
// Post fetched by the cutoff, how many such Posts there are, and how many
// Posts are new since. If no Post is shown, it also returns when gator last
// fetched the user's Feeds.
func (s *PostService) Queue(ctx context.Context, userID uuid.UUID, cutoff time.Time) (models.Queue, error) {
	count, err := s.CountQueueCreatedBy(ctx, userID, cutoff)
	if err != nil {
		return models.Queue{}, err
	}
	// The chip reveals what it counted: Posts fetched by the moment it counts.
	newCutoff := time.Now()
	newCount, err := s.Repo.CountQueuePostsCreatedAfter(ctx, database.CountQueuePostsCreatedAfterParams{
		UserID: userID.String(),
		Cutoff: cutoff.UTC(),
	})
	if err != nil {
		return models.Queue{}, fmt.Errorf("count new posts: %w", err)
	}
	queue := models.Queue{Count: count, Cutoff: cutoff, New: int(newCount), NewCutoff: newCutoff}

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

// SearchPosts returns the newest 100 Posts, in any Place, whose title or
// description contains query.
func (s *PostService) SearchPosts(ctx context.Context, userID uuid.UUID, query string) ([]models.Post, error) {
	dbPosts, err := s.Repo.SearchPostsByUser(ctx, database.SearchPostsByUserParams{
		UserID:     userID.String(),
		SearchText: query,
		LimitCount: 100,
	})
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(dbPosts))
	for _, dbPost := range dbPosts {
		posts = append(posts, models.Post{
			ID:          uuid.MustParse(dbPost.ID),
			Title:       dbPost.Title,
			Link:        dbPost.Url,
			Description: dbPost.Description.String,
			PublishedAt: dbPost.PublishedAt,
			FeedID:      uuid.MustParse(dbPost.FeedID),
			FeedName:    dbPost.FeedName,
		})
	}

	return posts, nil
}
