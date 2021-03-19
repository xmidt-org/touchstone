package touchhttp

import (
	"time"
)

// clock is a known now() implementation that can produce
// consistent, expected durations for metrics assertions.
type clock struct {
	start    time.Time
	duration time.Duration
	count    int
}

func (c *clock) Now() time.Time {
	c.count++
	if c.count > 1 {
		return c.start.Add(c.duration)
	}

	return c.start
}
