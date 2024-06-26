package limiter

import (
	"context"

	"golang.org/x/time/rate"
)

type MultiLimiter struct {
	limiters []RateLimiter
}

func (m *MultiLimiter) Wait(ctx context.Context) error {
	for _, l := range m.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiLimiter) Limit() rate.Limit {
	return m.limiters[0].Limit()
}
