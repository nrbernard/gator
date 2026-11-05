package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
)

// Place is where a Post is for a user. Every Post is in exactly one
// (docs/adr/0001-exclusive-post-states.md).
type Place string

const (
	PlaceQueue       Place = "queue"
	PlaceReadingList Place = "reading_list"
	PlaceArchive     Place = "archive"
	PlaceTrash       Place = "trash"
)

// PostState is a Post's Place for a user, where it came from if it is in the
// Trash, and its Read mark.
type PostState struct {
	Place       Place
	TrashedFrom Place
	Read        bool
}

var (
	// ErrInvalidMove is returned for a move that doesn't apply to the Post's
	// current Place, such as Keep on a Post already on the Reading list.
	ErrInvalidMove  = errors.New("invalid move")
	ErrPostNotFound = errors.New("post not found")
)

// PostStateService owns every move of a Post between Places, and rejects the
// moves that don't apply.
type PostStateService struct {
	Repo *database.Queries
}

// State returns where the Post is for the user.
func (s *PostStateService) State(ctx context.Context, userID, postID uuid.UUID) (PostState, error) {
	row, err := s.load(ctx, userID, postID)
	if err != nil {
		return PostState{}, err
	}
	return stateFromRow(row), nil
}

// Keep moves a Queue Post to the Reading list.
func (s *PostStateService) Keep(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "keep", []Place{PlaceQueue}, func(PostState) PostState {
		return PostState{Place: PlaceReadingList}
	})
}

// Trash moves a Post to the Trash, remembering where it came from.
func (s *PostStateService) Trash(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "trash", []Place{PlaceQueue, PlaceReadingList, PlaceArchive}, func(current PostState) PostState {
		return PostState{Place: PlaceTrash, TrashedFrom: current.Place, Read: current.Read}
	})
}

// move applies to to the Post's current state, provided the Post is in one of
// the from Places. The Read mark's original time survives a move that keeps it.
func (s *PostStateService) move(ctx context.Context, userID, postID uuid.UUID, name string, from []Place, to func(PostState) PostState) error {
	row, err := s.load(ctx, userID, postID)
	if err != nil {
		return err
	}
	current := stateFromRow(row)
	if !slices.Contains(from, current.Place) {
		return fmt.Errorf("%w: %s a post in %s", ErrInvalidMove, name, current.Place)
	}
	next := to(current)

	readAt := row.ReadAt
	switch {
	case !next.Read:
		readAt = sql.NullTime{}
	case !readAt.Valid:
		readAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	return s.Repo.SetPostState(ctx, database.SetPostStateParams{
		UserID:      userID.String(),
		PostID:      postID.String(),
		Place:       string(next.Place),
		TrashedFrom: sql.NullString{String: string(next.TrashedFrom), Valid: next.TrashedFrom != ""},
		ReadAt:      readAt,
	})
}

func (s *PostStateService) load(ctx context.Context, userID, postID uuid.UUID) (database.GetPostStateRow, error) {
	row, err := s.Repo.GetPostState(ctx, database.GetPostStateParams{UserID: userID.String(), PostID: postID.String()})
	if errors.Is(err, sql.ErrNoRows) {
		return row, fmt.Errorf("%w: %s", ErrPostNotFound, postID)
	}
	return row, err
}

func stateFromRow(row database.GetPostStateRow) PostState {
	if !row.Place.Valid {
		return PostState{Place: PlaceQueue}
	}
	return PostState{
		Place:       Place(row.Place.String),
		TrashedFrom: Place(row.TrashedFrom.String),
		Read:        row.ReadAt.Valid,
	}
}
