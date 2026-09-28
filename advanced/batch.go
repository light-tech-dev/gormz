package advanced

import (
	"context"
	"fmt"
	"sync"

	"github.com/light-tech-dev/gormz"
)

// ═══════════════════════════════════════════════
// Batch Processing — معالجة دفعية
// ═══════════════════════════════════════════════

// BatchConfig إعدادات المعالجة الدفعية.
type BatchConfig struct {
	BatchSize   int
	Workers     int
	StopOnError bool
}

// DefaultBatchConfig إعدادات افتراضية.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		BatchSize:   1000,
		Workers:     4,
		StopOnError: false,
	}
}

// ═══════════════════════════════════════════════
// BatchProcess — معالجة دفعية متوازية
// ═══════════════════════════════════════════════

// BatchResult نتيجة المعالجة.
type BatchResult struct {
	Total   int
	Success int
	Failed  int
	Errors  []error
}

// ProcessBatch يعالج slice بالتوازي.
//
//	users := []User{...}
//	result := advanced.ProcessBatch(ctx, users, advanced.DefaultBatchConfig(),
//	    func(batch []User) error {
//	        return gormz.New[User]().CreateMany(batch)
//	    })
func ProcessBatch[T any](ctx context.Context, items []T, cfg BatchConfig, fn func(batch []T) error) *BatchResult {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	result := &BatchResult{Total: len(items)}
	if len(items) == 0 {
		return result
	}

	// إنشاء دفعات
	var batches [][]T
	for i := 0; i < len(items); i += cfg.BatchSize {
		end := i + cfg.BatchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	batchCh := make(chan []T, len(batches))
	for _, b := range batches {
		batchCh <- b
	}
	close(batchCh)

	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := fn(batch); err != nil {
					mu.Lock()
					result.Failed += len(batch)
					result.Errors = append(result.Errors, err)
					mu.Unlock()

					if cfg.StopOnError {
						cancel()
						return
					}
				} else {
					mu.Lock()
					result.Success += len(batch)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return result
}

// ═══════════════════════════════════════════════
// StreamProcessing — معالجة streaming
// ═══════════════════════════════════════════════

// Stream يقرأ من DB ويعالج في streaming.
//
//	advanced.Stream[User](ctx, 1000, func(batch []User) error {
//	    return nil
//	})
func Stream[T any](ctx context.Context, batchSize int, fn func(batch []T) error) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	var offset int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		items, err := gormz.New[T]().
			Limit(batchSize).
			Offset(offset).
			All()
		if err != nil {
			return err
		}

		if len(items) == 0 {
			break
		}

		if err := fn(items); err != nil {
			return fmt.Errorf("gormz/advanced: stream batch at offset %d failed: %w", offset, err)
		}

		if len(items) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}

// ═══════════════════════════════════════════════
// Parallel Queries
// ═══════════════════════════════════════════════

// ParallelQuery يمثل استعلامًا متوازيًا.
type ParallelQuery struct {
	workers int
	tasks   []func() error
}

// Parallel ينشئ معالجًا متوازيًا.
//
//	p := advanced.Parallel()
//	p.Add(func() error { ... })
//	err := p.Run(ctx)
func Parallel() *ParallelQuery {
	return &ParallelQuery{
		workers: 4,
	}
}

// WithWorkers يحدد عدد الـ workers.
func (p *ParallelQuery) WithWorkers(n int) *ParallelQuery {
	if n > 0 {
		p.workers = n
	}
	return p
}

// Add يضيف مهمة.
func (p *ParallelQuery) Add(fn func() error) *ParallelQuery {
	p.tasks = append(p.tasks, fn)
	return p
}

// Run ينفّذ كل المهام.
func (p *ParallelQuery) Run(ctx context.Context) error {
	if len(p.tasks) == 0 {
		return nil
	}

	taskCh := make(chan func() error, len(p.tasks))
	for _, t := range p.tasks {
		taskCh <- t
	}
	close(taskCh)

	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := task(); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
						cancel()
					}
					mu.Unlock()
					return
				}
			}
		}()
	}

	wg.Wait()
	return firstErr
}