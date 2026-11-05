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

	if err := s.Trash(context.Background(), f.userID, postID, PlaceQueue); err != nil {
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

	if err := s.Trash(ctx, f.userID, postID, PlaceReadingList); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceTrash, TrashedFrom: PlaceReadingList})
}

func TestPostStateService_RejectsMovesThatDontApply(t *testing.T) {
	type move func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error
	keep := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.Keep(ctx, userID, postID)
	}
	trashFrom := func(from Place) move {
		return func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
			return s.Trash(ctx, userID, postID, from)
		}
	}
	trashed := PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue}

	tests := []struct {
		name    string
		prepare []move
		move    move
		want    PostState
	}{
		{"keep a Reading list post", []move{keep}, keep, PostState{Place: PlaceReadingList}},
		{"keep a trashed post", []move{trashFrom(PlaceQueue)}, keep, trashed},
		{"trash a trashed post", []move{trashFrom(PlaceQueue)}, trashFrom(PlaceTrash), trashed},
		{"trash a Reading list post seen in the Queue", []move{keep}, trashFrom(PlaceQueue), PostState{Place: PlaceReadingList}},
		{"trash a Queue post seen on the Reading list", nil, trashFrom(PlaceReadingList), PostState{Place: PlaceQueue}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, setupTestDB(t))
			s := &PostStateService{Repo: f.queries}
			ctx := context.Background()
			postID := f.post(t, "Post", time.Now())
			for _, m := range tc.prepare {
				if err := m(s, ctx, f.userID, postID); err != nil {
					t.Fatalf("prepare: %v", err)
				}
			}

			err := tc.move(s, ctx, f.userID, postID)

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
