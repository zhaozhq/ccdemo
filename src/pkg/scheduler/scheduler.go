package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"ccdemo/src/pkg/aggregator"
	"ccdemo/src/pkg/fetcher"
	"ccdemo/src/pkg/logger"
	"ccdemo/src/pkg/storage"
	"go.uber.org/zap"
)

// Notifier is the interface for sending notifications.
type Notifier interface {
	SendHotSearch(chatID int64, items []storage.HotSearch) error
}

// Scheduler orchestrates periodic fetching, aggregation, storage, and push.
type Scheduler struct {
	store    *storage.Storage
	fetchers []fetcher.Fetcher
	notifier Notifier
	chatID   int64
	cron     *cron.Cron
	agg      *aggregator.Aggregator
}

// New creates a new Scheduler.
func New(store *storage.Storage, fetchers []fetcher.Fetcher, notifier Notifier, chatID int64) *Scheduler {
	return &Scheduler{
		store:    store,
		fetchers: fetchers,
		notifier: notifier,
		chatID:   chatID,
		cron:     cron.New(),
		agg:      aggregator.New(),
	}
}

// Start starts the cron scheduler with the given cron spec.
// Adds two jobs:
//  1. Main job: FetchAndStore + Push at the given cron spec
//  2. Cleanup job: DeleteBefore 24h ago at "5 0 * * *"
func (s *Scheduler) Start(cronSpec string) error {
	_, err := s.cron.AddFunc(cronSpec, func() {
		ctx := context.Background()
		if err := s.FetchAndStore(ctx); err != nil {
			logger.Error("scheduled fetch and store failed", zap.Error(err))
			return
		}
		if err := s.Push(); err != nil {
			logger.Error("scheduled push failed", zap.Error(err))
		}
	})
	if err != nil {
		return fmt.Errorf("add main job: %w", err)
	}

	_, err = s.cron.AddFunc("5 0 * * *", func() {
		if err := s.Cleanup(); err != nil {
			logger.Error("scheduled cleanup failed", zap.Error(err))
		}
	})
	if err != nil {
		return fmt.Errorf("add cleanup job: %w", err)
	}

	s.cron.Start()
	return nil
}

// Stop stops the cron scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

// FetchAndStore fetches from all platforms in parallel, aggregates top 20, stores in DB.
func (s *Scheduler) FetchAndStore(ctx context.Context) error {
	if len(s.fetchers) == 0 {
		return fmt.Errorf("no fetchers configured")
	}

	results := make([]fetcher.FetchResult, 0, len(s.fetchers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, f := range s.fetchers {
		wg.Add(1)
		go func(fetcher fetcher.Fetcher) {
			defer wg.Done()
			res, err := fetcher.Fetch(ctx)
			if err != nil {
				logger.Warn("fetcher failed",
					zap.String("platform", fetcher.Name()),
					zap.Error(err),
				)
				return
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(f)
	}

	wg.Wait()

	if len(results) == 0 {
		return fmt.Errorf("all fetchers failed")
	}

	items := s.agg.Aggregate(results, 20)
	if err := s.store.Save(items); err != nil {
		return fmt.Errorf("save aggregated items: %w", err)
	}

	return nil
}

// Push sends current hot searches from DB to the chat.
func (s *Scheduler) Push() error {
	items, err := s.store.ListAll()
	if err != nil {
		return fmt.Errorf("list all hot searches: %w", err)
	}
	if err := s.notifier.SendHotSearch(s.chatID, items); err != nil {
		return fmt.Errorf("send hot search: %w", err)
	}
	return nil
}

// Cleanup removes data older than 24 hours.
func (s *Scheduler) Cleanup() error {
	cutoff := time.Now().Add(-24 * time.Hour)
	if err := s.store.DeleteBefore(cutoff); err != nil {
		return fmt.Errorf("delete before %v: %w", cutoff, err)
	}
	return nil
}
