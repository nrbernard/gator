package clock

import (
	"context"
	"errors"
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

func TestClock_SleepUntilWaitsForTheTime(t *testing.T) {
	c := New()
	until := time.Now().Add(20 * time.Millisecond)

	if err := c.SleepUntil(context.Background(), until); err != nil {
		t.Fatalf("SleepUntil: %v", err)
	}

	if now := time.Now(); now.Before(until) {
		t.Errorf("woke at %v, before %v", now, until)
	}
}

func TestClock_SleepUntilStopsWhenTheContextIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := New().SleepUntil(ctx, time.Now().Add(time.Hour))

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestClock_NullSleepsByMovingOnToTheTime(t *testing.T) {
	at := time.Date(2026, 11, 5, 9, 30, 15, 0, time.UTC)
	c := NewNull(at)

	if err := c.SleepUntil(context.Background(), at.Add(time.Second)); err != nil {
		t.Fatalf("SleepUntil: %v", err)
	}
	if err := c.SleepUntil(context.Background(), at); err != nil {
		t.Fatalf("SleepUntil the past: %v", err)
	}

	if got, want := c.Now(), at.Add(time.Second); !got.Equal(want) {
		t.Errorf("Now() = %v, want %v", got, want)
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
