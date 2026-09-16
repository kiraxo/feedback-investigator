package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kiraxo/feedback-investigator/backend/graph/model"
	"github.com/kiraxo/feedback-investigator/backend/internal/storage"
)

// Service executes investigations and persists completed runs.
type Service struct {
	repository storage.RunRepository
}

// NewService uses an in-memory repository by default.
// A PostgreSQL repository can be supplied by the server.
func NewService(
	repositories ...storage.RunRepository,
) *Service {
	var repository storage.RunRepository

	if len(repositories) > 0 && repositories[0] != nil {
		repository = repositories[0]
	} else {
		repository = storage.NewMemoryRepository()
	}

	return &Service{
		repository: repository,
	}
}

func (s *Service) Run(
	ctx context.Context,
	input model.RunInvestigationInput,
) (*model.AgentRun, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(input.Feedback) == 0 {
		return nil, errors.New("at least one feedback item is required")
	}

	if input.Mode != model.AgentModeDemo ||
		input.Provider != model.ModelProviderDemo {
		return nil, fmt.Errorf(
			"provider %s in %s mode is not configured yet; use DEMO mode with the DEMO provider",
			input.Provider,
			input.Mode,
		)
	}

	started := time.Now().UTC()

	run := &model.AgentRun{
		ID:            uuid.NewString(),
		Status:        model.AgentRunStatusRunning,
		Mode:          input.Mode,
		Provider:      input.Provider,
		FeedbackCount: int32(len(input.Feedback)),
		Results:       make([]*model.FeedbackResult, 0, len(input.Feedback)),
		Tasks:         make([]*model.InvestigationTask, 0),
		StartedAt:     started.Format(time.RFC3339Nano),
	}

	for index, item := range input.Feedback {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if item == nil {
			return nil, fmt.Errorf(
				"feedback item at index %d is null",
				index,
			)
		}

		if item.FeedbackID == "" {
			return nil, fmt.Errorf(
				"feedback item at index %d has no feedbackId",
				index,
			)
		}

		if item.FeedbackText == "" {
			return nil, fmt.Errorf(
				"feedback item %s has no feedbackText",
				item.FeedbackID,
			)
		}

		classification, err := classify(*item)
		if err != nil {
			return nil, fmt.Errorf(
				"classify feedback %s: %w",
				item.FeedbackID,
				err,
			)
		}

		run.Results = append(run.Results, &model.FeedbackResult{
			FeedbackID:     item.FeedbackID,
			FeedbackText:   item.FeedbackText,
			Product:        item.Product,
			Source:         item.Source,
			OccurredAt:     item.OccurredAt,
			Period:         item.Period,
			Classification: classification,
		})
	}

	run.Tasks = buildTasks(run.Results, started)

	completed := time.Now().UTC().Format(time.RFC3339Nano)
	run.CompletedAt = &completed
	run.Status = model.AgentRunStatusCompleted

	if err := s.repository.Save(ctx, run); err != nil {
		return nil, fmt.Errorf("save agent run: %w", err)
	}

	return run, nil
}

func (s *Service) GetRun(
	ctx context.Context,
	id string,
) (*model.AgentRun, bool, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) ListRuns(
	ctx context.Context,
	limit int,
) ([]*model.AgentRun, error) {
	return s.repository.List(ctx, limit)
}

func (s *Service) ListTasks(
	ctx context.Context,
	runID *string,
) ([]*model.InvestigationTask, error) {
	return s.repository.ListTasks(ctx, runID)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}

func (s *Service) Close() {
	s.repository.Close()
}
