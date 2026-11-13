package jobs

import (
	"context"
	"log"
	"time"
)

func (s *Store) StartJanitor(ctx context.Context, ttl, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n := s.DeleteExpired(ttl)
				if n > 0 {
					log.Printf("janitor: removed %d expired jobs", n)
				}
			}
		}
	}()
}
