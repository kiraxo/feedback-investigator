package storage

import (
	"context"

	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

// RunRepository stores completed agent runs.
// Implementations can use memory, PostgreSQL, or another database.
type RunRepository interface {
	Save(
		ctx context.Context,
		run *model.AgentRun,
	) error

	Get(
		ctx context.Context,
		id string,
	) (*model.AgentRun, bool, error)

	List(
		ctx context.Context,
		limit int,
	) ([]*model.AgentRun, error)

	ListTasks(
		ctx context.Context,
		runID *string,
	) ([]*model.InvestigationTask, error)

	Ping(ctx context.Context) error

	Close()
}
