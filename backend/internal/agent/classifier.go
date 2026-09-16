package agent

import (
	"fmt"
	"strings"

	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

func classify(input model.FeedbackInput) (*model.Classification, error) {
	originalText := input.FeedbackText
	evidence := strings.TrimSpace(originalText)

	if evidence == "" {
		return nil, fmt.Errorf("feedback text is empty")
	}

	normalized := strings.ToLower(evidence)

	topic := model.TopicNeutral
	severity := int32(0)
	requiresInvestigation := false

	switch {
	case containsAny(normalized,
		"late",
		"delayed",
		"delay",
		"not arrived",
		"arrived after",
		"missing shipment",
		"missing package",
	):
		topic = model.TopicDeliveryDelay
		severity = 2
		requiresInvestigation = true

		if containsAny(normalized,
			"event",
			"launch",
			"deadline",
			"four days",
			"five days",
			"week late",
		) {
			severity = 3
		}

	case containsAny(normalized,
		"damaged",
		"crushed",
		"bent",
		"broken",
		"torn",
		"wet package",
		"damaged package",
	):
		topic = model.TopicDamagedPackaging
		severity = 2
		requiresInvestigation = true

		if containsAny(normalized,
			"unusable",
			"destroyed",
			"completely damaged",
			"several",
			"multiple",
		) {
			severity = 3
		}

	case containsAny(normalized,
		"blurry",
		"blurred",
		"faded",
		"wrong color",
		"wrong colour",
		"pixelated",
		"unreadable",
		"print quality",
		"misprint",
	):
		topic = model.TopicPrintQuality
		severity = 2
		requiresInvestigation = true

		if containsAny(normalized,
			"unreadable",
			"completely",
			"cannot read",
			"can't read",
		) {
			severity = 3
		}

	case containsAny(normalized,
		"peel",
		"peeling",
		"adhesive",
		"not stick",
		"won't stick",
		"wont stick",
		"falling off",
	):
		topic = model.TopicAdhesion
		severity = 2
		requiresInvestigation = true

	case containsAny(normalized,
		"artwork",
		"proof",
		"revision",
		"design change",
		"missed change",
		"incorrect design",
	):
		topic = model.TopicArtworkProcess
		severity = 2
		requiresInvestigation = true

	case containsAny(normalized,
		"excellent",
		"great",
		"happy",
		"perfect",
		"love",
		"amazing",
		"arrived on time",
	):
		topic = model.TopicPositive
		severity = 0
		requiresInvestigation = false

	case containsAny(normalized,
		"problem",
		"issue",
		"poor",
		"bad",
		"wrong",
		"missing",
	):
		topic = model.TopicOther
		severity = 1
		requiresInvestigation = true

	default:
		topic = model.TopicNeutral
		severity = 0
		requiresInvestigation = false
	}

	evidenceValid := strings.Contains(originalText, evidence)

	if !evidenceValid {
		return nil, fmt.Errorf(
			"evidence was not found in original feedback %s",
			input.FeedbackID,
		)
	}

	return &model.Classification{
		Topic:                 topic,
		Severity:              severity,
		Evidence:              evidence,
		RequiresInvestigation: requiresInvestigation,
		EvidenceValid:         true,
	}, nil
}

func containsAny(text string, phrases ...string) bool {
	for _, phrase := range phrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}
