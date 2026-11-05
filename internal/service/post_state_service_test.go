package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
)

// fixture is a user following one Feed, for seeding Posts.
type fixture struct {
	queries *database.Queries
	userID  uuid.UUID
	feedID  uuid.UUID
}

func newFixture(t *testing.T, queries *database.Queries) fixture {
	t.Helper()
	ctx := context.Background()

	userID := uuid.New()
	if _, err := queries.CreateUser(ctx, database.CreateUserParams{ID: userID.String(), Name: "nick"}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	feedID := uuid.New()
	if _, err := queries.CreateFeed(ctx, database.CreateFeedParams{
		ID: feedID.String(), Name: "Test Feed", Url: "http://example.com/" + feedID.String(), UserID: userID.String(),
	}); err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if _, err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.New().String(), UserID: userID.String(), FeedID: feedID.String(),
	}); err != nil {
		t.Fatalf("create feed follow: %v", err)
	}
	return fixture{queries: queries, userID: userID, feedID: feedID}
}

func (f fixture) post(t *testing.T, title string, publishedAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := f.queries.CreatePost(context.Background(), database.CreatePostParams{
		ID: id.String(), Title: title, Url: "http://example.com/posts/" + id.String(), PublishedAt: publishedAt, FeedID: f.feedID.String(),
	}); err != nil {
		t.Fatalf("create post: %v", err)
	}
	return id
}

func assertState(t *testing.T, s *PostStateService, userID, postID uuid.UUID, want PostState) {
	t.Helper()
	got, err := s.State(context.Background(), userID, postID)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if got != want {
		t.Errorf("state = %+v, want %+v", got, want)
	}
}

func TestPostStateService_NewPostIsInTheQueue(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := &PostStateService{Repo: f.queries}
	postID := f.post(t, "Fresh", time.Now())

	assertState(t, s, f.userID, postID, PostState{Place: PlaceQueue})
}

func TestPostStateService_KeepMovesAQueuePostToTheReadingList(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := &PostStateService{Repo: f.queries}
	postID := f.post(t, "Worth reading", time.Now())

	if err := s.Keep(context.Background(), f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList})
}

func TestPostStateService_TrashMovesAQueuePostToTheTrashFromTheQueue(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := &PostStateService{Repo: f.queries}
	postID := f.post(t, "Not for me", time.Now())

	if err := s.Trash(context.Background(), f.userID, postID); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
}

func TestPostStateService_TrashMovesAReadingListPostToTheTrashFromTheReadingList(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := &PostStateService{Repo: f.queries}
	ctx := context.Background()
	postID := f.post(t, "Changed my mind", time.Now())
	if err := s.Keep(ctx, f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	if err := s.Trash(ctx, f.userID, postID); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceTrash, TrashedFrom: PlaceReadingList})
}

func TestPostStateService_RejectsMovesThatDontApply(t *testing.T) {
	tests := []struct {
		name    string
		prepare []string
		move    string
		want    PostState
	}{
		{"keep a Reading list post", []string{"keep"}, "keep", PostState{Place: PlaceReadingList}},
		{"keep a trashed post", []string{"trash"}, "keep", PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue}},
		{"trash a trashed post", []string{"trash"}, "trash", PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, setupTestDB(t))
			s := &PostStateService{Repo: f.queries}
			ctx := context.Background()
			postID := f.post(t, "Post", time.Now())
			moves := map[string]func(context.Context, uuid.UUID, uuid.UUID) error{"keep": s.Keep, "trash": s.Trash}
			for _, m := range tc.prepare {
				if err := moves[m](ctx, f.userID, postID); err != nil {
					t.Fatalf("%s: %v", m, err)
				}
			}

			err := moves[tc.move](ctx, f.userID, postID)

			if !errors.Is(err, ErrInvalidMove) {
				t.Errorf("err = %v, want ErrInvalidMove", err)
			}
			assertState(t, s, f.userID, postID, tc.want)
		})
	}
}

func TestPostStateService_RejectsMovesOfAMissingPost(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := &PostStateService{Repo: f.queries}

	err := s.Keep(context.Background(), f.userID, uuid.New())

	if !errors.Is(err, ErrPostNotFound) {
		t.Errorf("err = %v, want ErrPostNotFound", err)
	}
}
