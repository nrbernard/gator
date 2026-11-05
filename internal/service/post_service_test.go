package service

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostService_ReadingListIsUnreadFirstThenMostRecentlyKept(t *testing.T) {
	f := newFixture(t, setupTestDB(t))
	states := f.stateService()
	posts := &PostService{Repo: f.queries, Clock: f.clock}
	ctx := context.Background()
	now := time.Now()
	// Published order is the reverse of kept order, so it can't explain the result.
	first := f.post(t, "Kept first", now)
	second := f.post(t, "Kept second", now.Add(-1*time.Hour))
	third := f.post(t, "Kept third", now.Add(-2*time.Hour))
	fourth := f.post(t, "Kept fourth", now.Add(-3*time.Hour))
	for _, postID := range []uuid.UUID{first, second, third, fourth} {
		if err := states.Keep(ctx, f.userID, postID); err != nil {
			t.Fatalf("Keep: %v", err)
		}
		f.clock.Advance(time.Second)
	}
	// Reading a Post, or reading it and undoing that, doesn't change when it was kept.
	for _, postID := range []uuid.UUID{first, third, second} {
		if err := states.Open(ctx, f.userID, postID); err != nil {
			t.Fatalf("Open: %v", err)
		}
	}
	if err := states.ToggleRead(ctx, f.userID, second); err != nil {
		t.Fatalf("ToggleRead: %v", err)
	}

	list, err := posts.ReadingList(ctx, f.userID)
	if err != nil {
		t.Fatalf("ReadingList: %v", err)
	}

	var got []string
	for _, p := range list {
		got = append(got, p.Title)
	}
	want := []string{"Kept fourth", "Kept second", "Kept third", "Kept first"}
	if !slices.Equal(got, want) {
		t.Errorf("reading list = %q, want %q", got, want)
	}
}
