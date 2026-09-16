package agent

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

type issueGroup struct {
	product      string
	topic        model.Topic
	current      []*model.FeedbackResult
	previous     []*model.FeedbackResult
	highestLevel int32
}

func buildTasks(
	results []*model.FeedbackResult,
	createdAt time.Time,
) []*model.InvestigationTask {
	currentProductTotals := make(map[string]int)
	previousProductTotals := make(map[string]int)
	groups := make(map[string]*issueGroup)

	for _, result := range results {
		if result == nil || result.Classification == nil {
			continue
		}

		switch result.Period {
		case model.FeedbackPeriodCurrent:
			currentProductTotals[result.Product]++
		case model.FeedbackPeriodPrevious:
			previousProductTotals[result.Product]++
		}

		if !result.Classification.RequiresInvestigation {
			continue
		}

		key := result.Product + "\x00" + result.Classification.Topic.String()

		group, found := groups[key]
		if !found {
			group = &issueGroup{
				product:  result.Product,
				topic:    result.Classification.Topic,
				current:  make([]*model.FeedbackResult, 0),
				previous: make([]*model.FeedbackResult, 0),
			}
			groups[key] = group
		}

		switch result.Period {
		case model.FeedbackPeriodCurrent:
			group.current = append(group.current, result)

			if result.Classification.Severity > group.highestLevel {
				group.highestLevel = result.Classification.Severity
			}

		case model.FeedbackPeriodPrevious:
			group.previous = append(group.previous, result)
		}
	}

	tasks := make([]*model.InvestigationTask, 0)

	for _, group := range groups {
		// A task is created only when the same current-period issue
		// appears at least twice for the same product.
		if len(group.current) < 2 {
			continue
		}

		currentProductTotal := currentProductTotals[group.product]
		previousProductTotal := previousProductTotals[group.product]

		currentShare := percentage(
			len(group.current),
			currentProductTotal,
		)

		var previousShare *float64
		var change *float64
		comparisonStatus := "NO_PREVIOUS_PERIOD_DATA"

		if previousProductTotal > 0 {
			calculatedPreviousShare := percentage(
				len(group.previous),
				previousProductTotal,
			)

			calculatedChange := roundOne(
				currentShare - calculatedPreviousShare,
			)

			previousShare = &calculatedPreviousShare
			change = &calculatedChange

			switch {
			case calculatedChange > 0:
				comparisonStatus = "INCREASING"
			case calculatedChange < 0:
				comparisonStatus = "DECREASING"
			default:
				comparisonStatus = "UNCHANGED"
			}
		}

		priority := calculatePriority(
			len(group.current),
			group.highestLevel,
			currentShare,
		)

		urgent, urgentReason := calculateUrgency(
			priority,
			len(group.current),
			change,
		)

		evidence := make(
			[]*model.SupportingEvidence,
			0,
			len(group.current),
		)

		for _, result := range group.current {
			evidence = append(evidence, &model.SupportingEvidence{
				FeedbackID: result.FeedbackID,
				Text:       result.Classification.Evidence,
			})
		}

		task := &model.InvestigationTask{
			ID: uuid.NewString(),
			TaskKey: slug(group.product) + ":" +
				strings.ToLower(group.topic.String()),
			Title: fmt.Sprintf(
				"Investigate %s for %s",
				humanizeTopic(group.topic),
				group.product,
			),
			Product:  group.product,
			Topic:    group.topic,
			Priority: priority,
			Status:   model.TaskStatusOpen,
			Reason:   "Recurring issue detected in the current period.",
			Metrics: &model.InvestigationMetrics{
				CurrentIssueMessages:    int32(len(group.current)),
				CurrentProductMessages:  int32(currentProductTotal),
				CurrentSharePercentage:  currentShare,
				PreviousIssueMessages:   int32(len(group.previous)),
				PreviousProductMessages: int32(previousProductTotal),
				PreviousSharePercentage: previousShare,
				ChangePercentagePoints:  change,
				ComparisonStatus:        comparisonStatus,
			},
			SupportingEvidence: evidence,
			RecommendedAction:  recommendedAction(group.topic),
			Urgent:             urgent,
			UrgentReason:       urgentReason,
			CreatedAt:          createdAt.Format(time.RFC3339Nano),
		}

		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		leftRank := priorityRank(tasks[i].Priority)
		rightRank := priorityRank(tasks[j].Priority)

		if leftRank == rightRank {
			return tasks[i].TaskKey < tasks[j].TaskKey
		}

		return leftRank > rightRank
	})

	return tasks
}

func percentage(issueCount int, productTotal int) float64 {
	if productTotal == 0 {
		return 0
	}

	return roundOne(
		(float64(issueCount) / float64(productTotal)) * 100,
	)
}

func roundOne(value float64) float64 {
	return math.Round(value*10) / 10
}

func calculatePriority(
	currentIssueCount int,
	highestSeverity int32,
	currentShare float64,
) model.Priority {
	switch {
	case currentIssueCount >= 4 && highestSeverity == 3:
		return model.PriorityCritical

	case highestSeverity == 3:
		return model.PriorityHigh

	case currentIssueCount >= 3:
		return model.PriorityHigh

	case currentShare >= 40:
		return model.PriorityHigh

	default:
		return model.PriorityMedium
	}
}

func calculateUrgency(
	priority model.Priority,
	currentIssueCount int,
	change *float64,
) (bool, string) {
	if priority == model.PriorityCritical {
		return true, "Critical priority"
	}

	if priority == model.PriorityHigh {
		return true, "High priority"
	}

	if currentIssueCount >= 3 {
		return true, "Three or more current-period reports"
	}

	if change != nil && *change >= 10 {
		return true, "Issue share increased by at least 10 percentage points"
	}

	return false, "Standard investigation priority"
}

func priorityRank(priority model.Priority) int {
	switch priority {
	case model.PriorityCritical:
		return 4
	case model.PriorityHigh:
		return 3
	case model.PriorityMedium:
		return 2
	default:
		return 1
	}
}

func humanizeTopic(topic model.Topic) string {
	return strings.ToLower(
		strings.ReplaceAll(topic.String(), "_", " "),
	)
}

func slug(value string) string {
	lower := strings.ToLower(value)
	var builder strings.Builder
	previousDash := false

	for _, character := range lower {
		switch {
		case unicode.IsLetter(character) || unicode.IsDigit(character):
			builder.WriteRune(character)
			previousDash = false

		case !previousDash:
			builder.WriteRune('-')
			previousDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func recommendedAction(topic model.Topic) string {
	switch topic {
	case model.TopicDeliveryDelay:
		return "Compare promised delivery dates with dispatch and carrier tracking records before assigning a cause."

	case model.TopicDamagedPackaging:
		return "Review packaging photos, packing methods, and carrier handling records for the affected orders."

	case model.TopicPrintQuality:
		return "Compare the submitted artwork, approved proof, and production output for the affected orders."

	case model.TopicAdhesion:
		return "Review material, surface guidance, production batch data, and environmental conditions."

	case model.TopicArtworkProcess:
		return "Compare the customer request, approved proof, revision history, and final production artwork."

	default:
		return "Review the supporting reports, confirm the common cause, and assign an owner for corrective action."
	}
}
