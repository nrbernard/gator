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
	IsSaved     bool
	IsRead      bool
}
