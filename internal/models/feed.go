package models

import "github.com/google/uuid"

type Feed struct {
	ID   uuid.UUID
	Name string
	Url  string
}
