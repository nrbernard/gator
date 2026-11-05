package clock

import (
	"testing"
	"time"
)

func TestClock_TellsTheSystemTime(t *testing.T) {
	before := time.Now()
	got := New().Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestClock_NullStaysAtItsTimeUntilAdvanced(t *testing.T) {
	at := time.Date(2026, 11, 5, 9, 30, 15, 0, time.UTC)
	c := NewNull(at)

	for range 2 {
		if got := c.Now(); !got.Equal(at) {
			t.Errorf("Now() = %v, want %v", got, at)
		}
	}

	c.Advance(time.Minute)

	if got, want := c.Now(), time.Date(2026, 11, 5, 9, 31, 15, 0, time.UTC); !got.Equal(want) {
		t.Errorf("after Advance, Now() = %v, want %v", got, want)
	}
}
