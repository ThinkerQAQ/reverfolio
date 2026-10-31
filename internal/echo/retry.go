package echo

import (
	"context"
	"fmt"
	"time"
)

type apiStatusError struct {
	Code       int
	RetryAfter time.Duration
}

func (e apiStatusError) Error() string { return fmt.Sprintf("GitHub status %d", e.Code) }
func (g GitHub) get(ctx context.Context, path string, out any) error {
	for attempt := 0; attempt < 3; attempt++ {
		err := g.getOnce(ctx, path, out)
		if err == nil {
			return nil
		}
		status, ok := err.(apiStatusError)
		if !ok || !(status.Code == 429 || status.Code >= 500) || attempt == 2 {
			return err
		}
		delay := time.Duration(1<<attempt) * 250 * time.Millisecond
		if status.RetryAfter > 2*time.Second {
			return err
		}
		if status.RetryAfter > delay {
			delay = status.RetryAfter
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("GitHub retry limit reached")
}
