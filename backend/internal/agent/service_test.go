package agent

import (
	"context"
	"testing"

	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

func TestRunDemoCreatesRecurringInvestigationTask(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	input := model.RunInvestigationInput{
		Mode:     model.AgentModeDemo,
		Provider: model.ModelProviderDemo,
		Feedback: []*model.FeedbackInput{
			{
				FeedbackID:   "FB-001",
				FeedbackText: "My sticker order arrived four days late, and I needed it for an event.",
				Product:      "Die-cut stickers",
				Source:       "Demo review",
				OccurredAt:   "2026-09-10",
				Period:       model.FeedbackPeriodCurrent,
			},
			{
				FeedbackID:   "FB-002",
				FeedbackText: "The sticker package arrived two days late.",
				Product:      "Die-cut stickers",
				Source:       "Demo ticket",
				OccurredAt:   "2026-09-11",
				Period:       model.FeedbackPeriodCurrent,
			},
			{
				FeedbackID:   "FB-003",
				FeedbackText: "The stickers look excellent and arrived on time.",
				Product:      "Die-cut stickers",
				Source:       "Demo review",
				OccurredAt:   "2026-09-12",
				Period:       model.FeedbackPeriodCurrent,
			},
			{
				FeedbackID:   "FB-004",
				FeedbackText: "The previous sticker order arrived one day late.",
				Product:      "Die-cut stickers",
				Source:       "Demo review",
				OccurredAt:   "2026-09-06",
				Period:       model.FeedbackPeriodPrevious,
			},
			{
				FeedbackID:   "FB-005",
				FeedbackText: "The previous stickers looked great.",
				Product:      "Die-cut stickers",
				Source:       "Demo review",
				OccurredAt:   "2026-09-07",
				Period:       model.FeedbackPeriodPrevious,
			},
		},
	}

	run, err := service.Run(ctx, input)
	if err != nil {
		t.Fatalf("Run returned an unexpected error: %v", err)
	}

	if run.Status != model.AgentRunStatusCompleted {
		t.Fatalf(
			"expected COMPLETED status, got %s",
			run.Status,
		)
	}

	if run.FeedbackCount != 5 {
		t.Fatalf(
			"expected 5 feedback items, got %d",
			run.FeedbackCount,
		)
	}

	if len(run.Results) != 5 {
		t.Fatalf(
			"expected 5 classification results, got %d",
			len(run.Results),
		)
	}

	if len(run.Tasks) != 1 {
		t.Fatalf(
			"expected 1 recurring investigation task, got %d",
			len(run.Tasks),
		)
	}

	task := run.Tasks[0]

	if task.Topic != model.TopicDeliveryDelay {
		t.Errorf(
			"expected DELIVERY_DELAY topic, got %s",
			task.Topic,
		)
	}

	if task.Priority != model.PriorityHigh {
		t.Errorf(
			"expected HIGH priority, got %s",
			task.Priority,
		)
	}

	if !task.Urgent {
		t.Error("expected the high-priority task to be urgent")
	}

	if task.Metrics.CurrentIssueMessages != 2 {
		t.Errorf(
			"expected 2 current issue messages, got %d",
			task.Metrics.CurrentIssueMessages,
		)
	}

	if task.Metrics.CurrentProductMessages != 3 {
		t.Errorf(
			"expected 3 current product messages, got %d",
			task.Metrics.CurrentProductMessages,
		)
	}

	if task.Metrics.CurrentSharePercentage != 66.7 {
		t.Errorf(
			"expected current share 66.7, got %.1f",
			task.Metrics.CurrentSharePercentage,
		)
	}

	if task.Metrics.PreviousSharePercentage == nil {
		t.Fatal("expected a previous share percentage")
	}

	if *task.Metrics.PreviousSharePercentage != 50 {
		t.Errorf(
			"expected previous share 50.0, got %.1f",
			*task.Metrics.PreviousSharePercentage,
		)
	}

	if task.Metrics.ChangePercentagePoints == nil {
		t.Fatal("expected a percentage-point change")
	}

	if *task.Metrics.ChangePercentagePoints != 16.7 {
		t.Errorf(
			"expected change 16.7, got %.1f",
			*task.Metrics.ChangePercentagePoints,
		)
	}

	if task.Metrics.ComparisonStatus != "INCREASING" {
		t.Errorf(
			"expected INCREASING comparison, got %s",
			task.Metrics.ComparisonStatus,
		)
	}

	if len(task.SupportingEvidence) != 2 {
		t.Errorf(
			"expected 2 evidence records, got %d",
			len(task.SupportingEvidence),
		)
	}

	for _, result := range run.Results {
		if !result.Classification.EvidenceValid {
			t.Errorf(
				"expected valid evidence for %s",
				result.FeedbackID,
			)
		}
	}

	if run.CompletedAt == nil {
		t.Error("expected the run to have a completion time")
	}

	storedRun, found, err := service.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf(
			"GetRun returned an unexpected error: %v",
			err,
		)
	}

	if !found {
		t.Fatal("expected the completed run to be stored")
	}

	if storedRun.ID != run.ID {
		t.Errorf(
			"expected stored run %s, got %s",
			run.ID,
			storedRun.ID,
		)
	}
}

func TestPositiveFeedbackDoesNotCreateTask(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	run, err := service.Run(
		ctx,
		model.RunInvestigationInput{
			Mode:     model.AgentModeDemo,
			Provider: model.ModelProviderDemo,
			Feedback: []*model.FeedbackInput{
				{
					FeedbackID:   "FB-POSITIVE",
					FeedbackText: "The stickers look excellent and arrived on time.",
					Product:      "Die-cut stickers",
					Source:       "Demo review",
					OccurredAt:   "2026-09-12",
					Period:       model.FeedbackPeriodCurrent,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("Run returned an unexpected error: %v", err)
	}

	classification := run.Results[0].Classification

	if classification.Topic != model.TopicPositive {
		t.Errorf(
			"expected POSITIVE topic, got %s",
			classification.Topic,
		)
	}

	if classification.Severity != 0 {
		t.Errorf(
			"expected severity 0, got %d",
			classification.Severity,
		)
	}

	if classification.RequiresInvestigation {
		t.Error("positive feedback must not require investigation")
	}

	if len(run.Tasks) != 0 {
		t.Errorf(
			"expected no tasks for positive feedback, got %d",
			len(run.Tasks),
		)
	}
}

func TestRunRejectsEmptyFeedback(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	_, err := service.Run(
		ctx,
		model.RunInvestigationInput{
			Mode:     model.AgentModeDemo,
			Provider: model.ModelProviderDemo,
		},
	)

	if err == nil {
		t.Fatal("expected empty feedback to be rejected")
	}
}

func TestRunRejectsUnconfiguredLiveProvider(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	_, err := service.Run(
		ctx,
		model.RunInvestigationInput{
			Mode:     model.AgentModeLive,
			Provider: model.ModelProviderOpenai,
			Feedback: []*model.FeedbackInput{
				{
					FeedbackID:   "FB-001",
					FeedbackText: "The order arrived late.",
					Product:      "Die-cut stickers",
					Source:       "Demo review",
					OccurredAt:   "2026-09-10",
					Period:       model.FeedbackPeriodCurrent,
				},
			},
		},
	)

	if err == nil {
		t.Fatal("expected an unconfigured live provider to be rejected")
	}
}
