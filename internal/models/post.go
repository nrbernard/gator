package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Description string
	Link        string
	Title       string
	PublishedAt time.Time
	FeedID      uuid.UUID
	FeedName    string
}
