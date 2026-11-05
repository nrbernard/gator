// Package clock wraps the time of day, so tests can stop and move it.
package clock

import (
	"context"
	"time"
)

// Clock tells the current time: the system's, or, for a null clock, a
// stopped time that only Advance moves.
type Clock struct {
	null bool
	at   time.Time
}

// New returns the system clock.
func New() *Clock {
	return &Clock{}
}

// NewNull returns a clock stopped at now, for tests.
func NewNull(now time.Time) *Clock {
	return &Clock{null: true, at: now}
}

// Now is the current time.
func (c *Clock) Now() time.Time {
	if c.null {
		return c.at
	}
	return time.Now()
}

// SleepUntil returns once the clock reaches t, or when ctx is done. A null
// clock moves straight on to t instead of waiting.
func (c *Clock) SleepUntil(ctx context.Context, t time.Time) error {
	if c.null {
		if t.After(c.at) {
			c.at = t
		}
		return nil
	}
	wait := time.Until(t)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Advance moves a null clock on by d. The system clock can't be moved.
func (c *Clock) Advance(d time.Duration) {
	if !c.null {
		panic("clock: Advance on the system clock")
	}
	c.at = c.at.Add(d)
}
