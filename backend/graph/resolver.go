package graph

import (
	"github.com/kiraxo/feedback-investigator/backend/internal/agent"
)

// Resolver contains the application dependencies used by GraphQL.
type Resolver struct {
	AgentService *agent.Service
}

func NewResolver() *Resolver {
	return &Resolver{
		AgentService: agent.NewService(),
	}
}
