package models

import "github.com/google/uuid"

type Feed struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Url         string
}
