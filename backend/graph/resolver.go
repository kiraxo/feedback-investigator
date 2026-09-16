package graph

import (
	"github.com/kiraxo/feedback-investigator/backend/internal/agent"
)

// Resolver contains the application dependencies used by GraphQL.
type Resolver struct {
	AgentService *agent.Service
}

func NewResolver(
	agentService *agent.Service,
) *Resolver {
	if agentService == nil {
		agentService = agent.NewService()
	}

	return &Resolver{
		AgentService: agentService,
	}
}
