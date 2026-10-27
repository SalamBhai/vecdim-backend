package pipeline

import (
	"context"
	"log"
	"time"
)

// CronScheduler manages periodic background execution of the staged analysis pipeline.
type CronScheduler struct {
	runner   *Runner
	interval time.Duration
	stopChan chan struct{}
}

// NewCronScheduler initializes a background cron worker with a configurable interval.
func NewCronScheduler(runner *Runner, interval time.Duration) *CronScheduler {
	if interval <= 0 {
		interval = 15 * time.Minute // Default: check every 15 minutes
	}
	return &CronScheduler{
		runner:   runner,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins the background cron ticker loop.
func (c *CronScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	log.Printf("[Cron Scheduler] Background pipeline worker started (interval: %v)", c.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("[Cron Scheduler] Ticker fired: checking for UNPROCESSED staged records...")
				result, err := c.runner.TriggerPipeline(ctx, "cron_scheduler")
				if err != nil {
					log.Printf("[Cron Scheduler] Error running scheduled pipeline: %v", err)
				} else if result.UnprocessedCount > 0 {
					log.Printf("[Cron Scheduler] Processed %d staged records in %d ms",
						result.UnprocessedCount, result.ExecutionTimeMs)
				}
			case <-c.stopChan:
				ticker.Stop()
				log.Println("[Cron Scheduler] Background worker stopped.")
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop cleanly terminates the cron worker.
func (c *CronScheduler) Stop() {
	close(c.stopChan)
}
