package storage

import (
	"context"
	"sync"

	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

// MemoryRepository provides a zero-configuration repository for demo mode
// and automated tests.
type MemoryRepository struct {
	mu    sync.RWMutex
	runs  map[string]*model.AgentRun
	order []string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		runs:  make(map[string]*model.AgentRun),
		order: make([]string, 0),
	}
}

func (r *MemoryRepository) Save(
	ctx context.Context,
	run *model.AgentRun,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.runs[run.ID]; !exists {
		r.order = append([]string{run.ID}, r.order...)
	}

	r.runs[run.ID] = run
	return nil
}

func (r *MemoryRepository) Get(
	ctx context.Context,
	id string,
) (*model.AgentRun, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	run, found := r.runs[id]
	return run, found, nil
}

func (r *MemoryRepository) List(
	ctx context.Context,
	limit int,
) ([]*model.AgentRun, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	if limit > len(r.order) {
		limit = len(r.order)
	}

	runs := make([]*model.AgentRun, 0, limit)

	for _, id := range r.order[:limit] {
		runs = append(runs, r.runs[id])
	}

	return runs, nil
}

func (r *MemoryRepository) ListTasks(
	ctx context.Context,
	runID *string,
) ([]*model.InvestigationTask, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if runID != nil {
		run, found := r.runs[*runID]
		if !found {
			return []*model.InvestigationTask{}, nil
		}

		return run.Tasks, nil
	}

	tasks := make([]*model.InvestigationTask, 0)

	for _, id := range r.order {
		tasks = append(tasks, r.runs[id].Tasks...)
	}

	return tasks, nil
}

func (r *MemoryRepository) Ping(ctx context.Context) error {
	return ctx.Err()
}

func (r *MemoryRepository) Close() {
	// No resources need to be released.
}
