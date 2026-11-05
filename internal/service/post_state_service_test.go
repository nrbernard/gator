package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/clock"
	"github.com/nrbernard/gator/internal/database"
)

// fixture is a user following one Feed, for seeding Posts.
type fixture struct {
	queries *database.Queries
	userID  uuid.UUID
	feedID  uuid.UUID
	// clock is stopped at the time the fixture was made.
	clock *clock.Clock
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
	return fixture{queries: queries, userID: userID, feedID: feedID, clock: clock.NewNull(time.Now())}
}

// post adds a Post published at publishedAt that gator fetched at the
// fixture clock's time.
func (f fixture) post(t *testing.T, title string, publishedAt time.Time) uuid.UUID {
	t.Helper()
	return f.postFetchedAt(t, title, publishedAt, f.clock.Now())
}

// postFetchedAt adds a Post published at publishedAt that gator fetched at
// fetchedAt.
func (f fixture) postFetchedAt(t *testing.T, title string, publishedAt, fetchedAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := f.queries.CreatePost(context.Background(), database.CreatePostParams{
		ID: id.String(), Title: title, Url: "http://example.com/posts/" + id.String(), PublishedAt: publishedAt, FeedID: f.feedID.String(),
		CreatedAt: fetchedAt.UTC().Truncate(time.Second),
	}); err != nil {
		t.Fatalf("create post: %v", err)
	}
	return id
}

// stateService is the state service over the fixture's database and clock.
func (f fixture) stateService() *PostStateService {
	return &PostStateService{Repo: f.queries, Clock: f.clock}
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
	s := f.stateService()
	postID := f.post(t, "Fresh", time.Now())

	assertState(t, s, f.userID, postID, PostState{Place: PlaceQueue})
}

func TestPostStateService_KeepMovesAQueuePostToTheReadingList(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	postID := f.post(t, "Worth reading", time.Now())

	if err := s.Keep(context.Background(), f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList})
}

func TestPostStateService_TrashMovesAQueuePostToTheTrashFromTheQueue(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	postID := f.post(t, "Not for me", time.Now())

	if err := s.Trash(context.Background(), f.userID, postID, PlaceQueue); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
}

func TestPostStateService_TrashMovesAReadingListPostToTheTrashFromTheReadingList(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
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

func TestPostStateService_OpenMarksAReadingListPostReadAndLeavesItThere(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	postID := f.post(t, "Reading this now", time.Now())
	if err := s.Keep(ctx, f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	if err := s.Open(ctx, f.userID, postID); err != nil {
		t.Fatalf("Open: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList, Read: true})
}

func TestPostStateService_OpenKeepsAQueuePostAndMarksItRead(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	postID := f.post(t, "Clicked through", time.Now())

	if err := s.Open(context.Background(), f.userID, postID); err != nil {
		t.Fatalf("Open: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList, Read: true})
}

func TestPostStateService_ToggleReadFlipsTheMarkOnAReadingListPost(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	postID := f.post(t, "Opened by accident", time.Now())
	if err := s.Keep(ctx, f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	if err := s.ToggleRead(ctx, f.userID, postID); err != nil {
		t.Fatalf("ToggleRead: %v", err)
	}
	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList, Read: true})

	if err := s.ToggleRead(ctx, f.userID, postID); err != nil {
		t.Fatalf("ToggleRead: %v", err)
	}
	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList})
}

func TestPostStateService_ArchiveMovesAReadingListPostToTheArchive(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	postID := f.post(t, "Finished with it", time.Now())
	if err := s.Open(ctx, f.userID, postID); err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := s.Archive(ctx, f.userID, postID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceArchive, Read: true})
}

func TestPostStateService_UnarchiveMovesAnArchivedPostBackToTheReadingList(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	postID := f.post(t, "Worth rereading", time.Now())
	if err := s.Open(ctx, f.userID, postID); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Archive(ctx, f.userID, postID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	if err := s.Unarchive(ctx, f.userID, postID); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceReadingList, Read: true})
}

func TestPostStateService_TrashMovesAnArchivedPostToTheTrashFromTheArchive(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	postID := f.post(t, "Not worth keeping", time.Now())
	if err := s.Keep(ctx, f.userID, postID); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	if err := s.Archive(ctx, f.userID, postID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	if err := s.Trash(ctx, f.userID, postID, PlaceArchive); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	assertState(t, s, f.userID, postID, PostState{Place: PlaceTrash, TrashedFrom: PlaceArchive})
}

func TestPostStateService_RestoreReturnsATrashedPostWhereItCameFrom(t *testing.T) {
	type move func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error
	tests := []struct {
		name    string
		prepare []move
		want    PostState
	}{
		{"from the Queue", []move{
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Trash(ctx, userID, postID, PlaceQueue)
			},
		}, PostState{Place: PlaceQueue}},
		{"from the Reading list, keeping its Read mark", []move{
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Open(ctx, userID, postID)
			},
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Trash(ctx, userID, postID, PlaceReadingList)
			},
		}, PostState{Place: PlaceReadingList, Read: true}},
		{"from the Reading list, unread", []move{
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Keep(ctx, userID, postID)
			},
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Trash(ctx, userID, postID, PlaceReadingList)
			},
		}, PostState{Place: PlaceReadingList}},
		{"from the Archive", []move{
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Keep(ctx, userID, postID)
			},
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Archive(ctx, userID, postID)
			},
			func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
				return s.Trash(ctx, userID, postID, PlaceArchive)
			},
		}, PostState{Place: PlaceArchive}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, setupTestDB(t))
			s := f.stateService()
			ctx := context.Background()
			postID := f.post(t, "Trashed by mistake", time.Now())
			for _, m := range tc.prepare {
				if err := m(s, ctx, f.userID, postID); err != nil {
					t.Fatalf("prepare: %v", err)
				}
			}

			if err := s.Restore(ctx, f.userID, postID); err != nil {
				t.Fatalf("Restore: %v", err)
			}

			assertState(t, s, f.userID, postID, tc.want)
		})
	}
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
	open := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.Open(ctx, userID, postID)
	}
	toggleRead := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.ToggleRead(ctx, userID, postID)
	}
	archive := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.Archive(ctx, userID, postID)
	}
	unarchive := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.Unarchive(ctx, userID, postID)
	}
	restore := func(s *PostStateService, ctx context.Context, userID, postID uuid.UUID) error {
		return s.Restore(ctx, userID, postID)
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
		{"open a trashed post", []move{trashFrom(PlaceQueue)}, open, trashed},
		{"toggle read on a Queue post", nil, toggleRead, PostState{Place: PlaceQueue}},
		{"toggle read on a trashed post", []move{trashFrom(PlaceQueue)}, toggleRead, trashed},
		{"archive a Queue post", nil, archive, PostState{Place: PlaceQueue}},
		{"archive an archived post", []move{keep, archive}, archive, PostState{Place: PlaceArchive}},
		{"archive a trashed post", []move{trashFrom(PlaceQueue)}, archive, trashed},
		{"unarchive a Queue post", nil, unarchive, PostState{Place: PlaceQueue}},
		{"unarchive a Reading list post", []move{keep}, unarchive, PostState{Place: PlaceReadingList}},
		{"trash an archived post seen on the Reading list", []move{keep, archive}, trashFrom(PlaceReadingList), PostState{Place: PlaceArchive}},
		{"restore a Queue post", nil, restore, PostState{Place: PlaceQueue}},
		{"restore a Reading list post", []move{open}, restore, PostState{Place: PlaceReadingList, Read: true}},
		{"restore an archived post", []move{keep, archive}, restore, PostState{Place: PlaceArchive}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, setupTestDB(t))
			s := f.stateService()
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

// A removed Feed's Posts are out of the Queue, so a Post restored there would
// be lost; one trashed from elsewhere still goes back.
func TestPostStateService_RestoreOfARemovedFeedsPostOnlyLeavesTheTrashIfNotForTheQueue(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	fromQueue := f.post(t, "Trashed from the Queue", time.Now())
	fromReadingList := f.post(t, "Trashed from the Reading list", time.Now())
	if err := s.Trash(ctx, f.userID, fromQueue, PlaceQueue); err != nil {
		t.Fatalf("Trash: %v", err)
	}
	if err := s.Keep(ctx, f.userID, fromReadingList); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	if err := s.Trash(ctx, f.userID, fromReadingList, PlaceReadingList); err != nil {
		t.Fatalf("Trash: %v", err)
	}
	if err := (&FeedService{Repo: f.queries}).RemoveFeed(ctx, f.feedID); err != nil {
		t.Fatalf("RemoveFeed: %v", err)
	}

	if err := s.Restore(ctx, f.userID, fromQueue); !errors.Is(err, ErrInvalidMove) {
		t.Errorf("restore to the Queue: err = %v, want ErrInvalidMove", err)
	}
	if err := s.Restore(ctx, f.userID, fromReadingList); err != nil {
		t.Errorf("restore to the Reading list: %v", err)
	}

	assertState(t, s, f.userID, fromQueue, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
	assertState(t, s, f.userID, fromReadingList, PostState{Place: PlaceReadingList})
}

func TestPostStateService_RejectsMovesOfAMissingPost(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()

	err := s.Keep(context.Background(), f.userID, uuid.New())

	if !errors.Is(err, ErrPostNotFound) {
		t.Errorf("err = %v, want ErrPostNotFound", err)
	}
}

func TestPostStateService_TrashAllTrashesQueuePostsCreatedByTheCutoff(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	s := f.stateService()
	ctx := context.Background()
	cutoff := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	now := time.Now()
	seen := f.postFetchedAt(t, "Seen", now, cutoff.Add(-time.Hour))
	atCutoff := f.postFetchedAt(t, "Fetched as the page loaded", now, cutoff)
	unseen := f.postFetchedAt(t, "Fetched after the page loaded", now, cutoff.Add(time.Minute))
	kept := f.postFetchedAt(t, "Kept", now, cutoff.Add(-time.Hour))
	if err := s.Keep(ctx, f.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	if err := s.TrashAll(ctx, f.userID, cutoff); err != nil {
		t.Fatalf("TrashAll: %v", err)
	}

	assertState(t, s, f.userID, seen, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
	assertState(t, s, f.userID, atCutoff, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
	assertState(t, s, f.userID, unseen, PostState{Place: PlaceQueue})
	assertState(t, s, f.userID, kept, PostState{Place: PlaceReadingList})
}
