package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID          uuid.UUID
	Description string
	Link        string
	Title       string
	PublishedAt time.Time
	FeedID      uuid.UUID
	FeedName    string
	IsRead      bool
	// Restorable is whether Restore can take the Post out of the Trash: not
	// when it came from the Queue and its Feed has since been removed.
	Restorable bool
}

// Queue is the top of the user's Queue: the Post to triage next, if any, and
// how many Posts are in the Queue, that one included. Only Posts fetched by
// the cutoff are shown; the rest are new.
type Queue struct {
	Next  *Post
	Count int
	// Cutoff is the moment the Queue page loaded, or the chip last moved it.
	// The page sends it with each Queue request, and trash all spares Posts
	// fetched after it.
	Cutoff time.Time
	// New is how many Queue Posts gator fetched after the cutoff, counted at
	// NewCutoff. Tapping the chip moves the cutoff to NewCutoff.
	New       int
	NewCutoff time.Time
	// LastFetchedAt is when gator last fetched any of the user's Feeds, set
	// for an empty Queue; nil if it never has.
	LastFetchedAt *time.Time
}

// CutoffParam formats the cutoff for the page to send back.
func (q Queue) CutoffParam() string {
	return q.Cutoff.UTC().Format(time.RFC3339Nano)
}

// NewCutoffParam formats the cutoff the chip moves the page to.
func (q Queue) NewCutoffParam() string {
	return q.NewCutoff.UTC().Format(time.RFC3339Nano)
}
