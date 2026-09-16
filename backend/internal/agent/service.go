package agent

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

// Service executes investigations and keeps completed runs in memory.
// PostgreSQL will replace this in-memory storage in the next phase.
type Service struct {
	mu    sync.RWMutex
	runs  map[string]*model.AgentRun
	order []string
}

func NewService() *Service {
	return &Service{
		runs:  make(map[string]*model.AgentRun),
		order: make([]string, 0),
	}
}

func (s *Service) Run(input model.RunInvestigationInput) (*model.AgentRun, error) {
	if len(input.Feedback) == 0 {
		return nil, errors.New("at least one feedback item is required")
	}

	if input.Mode != model.AgentModeDemo || input.Provider != model.ModelProviderDemo {
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
		if item == nil {
			return nil, fmt.Errorf("feedback item at index %d is null", index)
		}

		if item.FeedbackID == "" {
			return nil, fmt.Errorf("feedback item at index %d has no feedbackId", index)
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

	s.mu.Lock()
	s.runs[run.ID] = run
	s.order = append([]string{run.ID}, s.order...)
	s.mu.Unlock()

	return run, nil
}

func (s *Service) GetRun(id string) (*model.AgentRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, found := s.runs[id]
	return run, found
}

func (s *Service) ListRuns(limit int) []*model.AgentRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	if limit > len(s.order) {
		limit = len(s.order)
	}

	runs := make([]*model.AgentRun, 0, limit)

	for _, id := range s.order[:limit] {
		runs = append(runs, s.runs[id])
	}

	return runs
}

func (s *Service) ListTasks(runID *string) []*model.InvestigationTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if runID != nil {
		run, found := s.runs[*runID]
		if !found {
			return []*model.InvestigationTask{}
		}

		return run.Tasks
	}

	tasks := make([]*model.InvestigationTask, 0)

	for _, id := range s.order {
		tasks = append(tasks, s.runs[id].Tasks...)
	}

	return tasks
}
