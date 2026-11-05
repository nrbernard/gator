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

// Label is the Place's name as the user sees it.
func (p Place) Label() string {
	switch p {
	case PlaceQueue:
		return "Queue"
	case PlaceReadingList:
		return "Reading list"
	case PlaceArchive:
		return "Archive"
	case PlaceTrash:
		return "Trash"
	default:
		return string(p)
	}
}

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

// Trash moves a Post from the Place the user saw it in to the Trash,
// remembering that Place. It is rejected if the Post is no longer there, so a
// stale screen can't trash a Post that has since moved.
func (s *PostStateService) Trash(ctx context.Context, userID, postID uuid.UUID, from Place) error {
	if from == PlaceTrash {
		return fmt.Errorf("%w: trash a post in %s", ErrInvalidMove, from)
	}
	return s.move(ctx, userID, postID, "trash", []Place{from}, func(current PostState) PostState {
		return PostState{Place: PlaceTrash, TrashedFrom: current.Place, Read: current.Read}
	})
}

// TrashAll trashes, from the Queue, every Queue Post gator fetched at or before
// cutoff, the moment the user's Queue page loaded, so Posts the user hasn't
// been shown are spared. It returns how many Posts it trashed.
func (s *PostStateService) TrashAll(ctx context.Context, userID uuid.UUID, cutoff time.Time) (int, error) {
	n, err := s.Repo.TrashQueuePostsCreatedBy(ctx, database.TrashQueuePostsCreatedByParams{
		UserID:    userID.String(),
		Cutoff:    cutoff.UTC(),
		ChangedAt: time.Now().UTC(),
	})
	if err != nil {
		return 0, fmt.Errorf("trash all: %w", err)
	}
	return int(n), nil
}

// Open records that the user opened the Post: it Keeps a Queue Post and marks
// it Read, and marks a Reading list Post Read, leaving it on the Reading list.
func (s *PostStateService) Open(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "open", []Place{PlaceQueue, PlaceReadingList}, func(PostState) PostState {
		return PostState{Place: PlaceReadingList, Read: true}
	})
}

// ToggleRead marks a Reading list Post Read, or clears its Read mark, so an
// accidental Open is easy to undo.
func (s *PostStateService) ToggleRead(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "toggle read", []Place{PlaceReadingList}, func(current PostState) PostState {
		return PostState{Place: PlaceReadingList, Read: !current.Read}
	})
}

// Archive moves a Reading list Post, finished with, to the Archive.
func (s *PostStateService) Archive(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "archive", []Place{PlaceReadingList}, func(current PostState) PostState {
		return PostState{Place: PlaceArchive, Read: current.Read}
	})
}

// Unarchive moves an archived Post back to the Reading list, to reread it.
func (s *PostStateService) Unarchive(ctx context.Context, userID, postID uuid.UUID) error {
	return s.move(ctx, userID, postID, "unarchive", []Place{PlaceArchive}, func(current PostState) PostState {
		return PostState{Place: PlaceReadingList, Read: current.Read}
	})
}

// Restore returns a trashed Post to the Place it was trashed from, with its
// Read mark. It is rejected for a Post that isn't in the Trash.
func (s *PostStateService) Restore(ctx context.Context, userID, postID uuid.UUID) error {
	current, err := s.State(ctx, userID, postID)
	if err != nil {
		return err
	}
	// Only a trashed Post has TrashedFrom, so move rejects every other Post.
	if current.TrashedFrom != PlaceQueue {
		return s.move(ctx, userID, postID, "restore", []Place{PlaceTrash}, func(current PostState) PostState {
			return PostState{Place: current.TrashedFrom, Read: current.Read}
		})
	}
	n, err := s.Repo.RestoreTrashedPostToQueue(ctx, database.RestoreTrashedPostToQueueParams{
		UserID: userID.String(),
		PostID: postID.String(),
	})
	if err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("%w: restore a post no longer in the trash", ErrInvalidMove)
	}
	return nil
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

	// UTC, so stored times sort as text alongside SQLite's CURRENT_TIMESTAMP.
	now := time.Now().UTC()
	readAt := row.ReadAt
	switch {
	case !next.Read:
		readAt = sql.NullTime{}
	case !readAt.Valid:
		readAt = sql.NullTime{Time: now, Valid: true}
	}

	return s.Repo.SetPostState(ctx, database.SetPostStateParams{
		UserID:      userID.String(),
		PostID:      postID.String(),
		Place:       string(next.Place),
		TrashedFrom: sql.NullString{String: string(next.TrashedFrom), Valid: next.TrashedFrom != ""},
		ReadAt:      readAt,
		ChangedAt:   now,
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
