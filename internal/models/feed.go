package models

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Url         string
	// LastFetchedAt is when gator last fetched the Feed successfully; nil if never.
	LastFetchedAt *time.Time
	// LastFetchError is why the Feed's last fetch failed, empty if it didn't.
	LastFetchError   string
	LastFetchErrorAt *time.Time
}
