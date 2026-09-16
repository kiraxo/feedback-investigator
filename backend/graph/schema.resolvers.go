package graph

import (
	"context"

	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

// RunInvestigation executes a complete feedback investigation.
func (r *mutationResolver) RunInvestigation(
	ctx context.Context,
	input model.RunInvestigationInput,
) (*model.AgentRun, error) {
	return r.AgentService.Run(ctx, input)
}

// Health reports whether the API and its configured repository are available.
func (r *queryResolver) Health(
	ctx context.Context,
) (*model.Health, error) {
	status := "ok"

	if err := r.AgentService.Ping(ctx); err != nil {
		status = "degraded"
	}

	return &model.Health{
		Status:  status,
		Service: "feedback-investigator-api",
		Version: "0.2.0",
	}, nil
}

// AgentRun returns one completed agent run by ID.
func (r *queryResolver) AgentRun(
	ctx context.Context,
	id string,
) (*model.AgentRun, error) {
	run, found, err := r.AgentService.GetRun(ctx, id)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return run, nil
}

// AgentRuns returns recent agent runs, newest first.
func (r *queryResolver) AgentRuns(
	ctx context.Context,
	limit *int32,
) ([]*model.AgentRun, error) {
	selectedLimit := 20

	if limit != nil {
		selectedLimit = int(*limit)
	}

	if selectedLimit > 100 {
		selectedLimit = 100
	}

	return r.AgentService.ListRuns(ctx, selectedLimit)
}

// InvestigationTasks returns tasks from one run or all stored runs.
func (r *queryResolver) InvestigationTasks(
	ctx context.Context,
	runID *string,
) ([]*model.InvestigationTask, error) {
	return r.AgentService.ListTasks(ctx, runID)
}

// Mutation returns the MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query returns the QueryResolver implementation.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
