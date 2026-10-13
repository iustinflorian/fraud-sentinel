package processor

import (
	"time"

	"fraud-sentinel/pkg/models"
)

func EvaluateFraudRisk(tx *models.Transaction) models.FraudResult {
	score := 0
	reasons := []string{}

	if tx.Amount > 10000 {
		score += 50
		reasons = append(reasons, "High amount threshold exceeded (> €10,000)")
	}

	if tx.Currency == "EUR" && tx.Location != "" && tx.Location != "EU" {
		score += 35
		reasons = append(reasons, "Currency mismatch with geographic location")
	}

	flagged := score >= 50

	return models.FraudResult{
		RiskScore:      score,
		Flagged:        flagged,
		FlaggedReasons: reasons,
		EvaluatedAt:    time.Now().UTC(),
	}
}
