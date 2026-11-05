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
}

// Queue is the top of the user's Queue: the Post to triage next, if any, and
// how many Posts are in the Queue, that one included.
type Queue struct {
	Next  *Post
	Count int
	// Cutoff is the moment the Queue page loaded. The page sends it with each
	// Queue request, and trash all spares Posts fetched after it.
	Cutoff time.Time
}

// CutoffParam formats the cutoff for the page to send back.
func (q Queue) CutoffParam() string {
	return q.Cutoff.UTC().Format(time.RFC3339Nano)
}
