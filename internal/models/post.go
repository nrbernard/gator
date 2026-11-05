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
}
