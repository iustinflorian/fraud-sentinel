package models

import "time"

type TransactionStatus string

const (
	StatusPending  TransactionStatus = "PENDING"
	StatusApproved TransactionStatus = "APPROVED"
	StatusFlagged  TransactionStatus = "FLAGGED"
	StatusRejected TransactionStatus = "REJECTED"
)

type Transaction struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	AccountID string    `json:"account_id" bson:"account_id"`
	Amount    float64   `json:"amount" bson:"amount"`
	Currency  string    `json:"currency" bson:"currency"`
	Merchant  string    `json:"merchant" bson:"merchant"`
	Location  string    `json:"location" bson:"location"` // e.g., "US", "FR", "JP"
	IPAddress string    `json:"ip_address" bson:"ip_address"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type FraudResult struct {
	RiskScore      int       `json:"risk_score" bson:"risk_score"`
	Flagged        bool      `json:"flagged" bson:"flagged"`
	FlaggedReasons []string  `json:"flagged_reasons,omitempty" bson:"flagged_reasons,omitempty"`
	EvaluatedAt    time.Time `json:"evaluated_at" bson:"evaluated_at"`
}

type TransactionRecord struct {
	ID          string            `json:"id" bson:"_id"`
	AccountID   string            `json:"account_id" bson:"account_id"`
	Amount      float64           `json:"amount" bson:"amount"`
	Currency    string            `json:"currency" bson:"currency"`
	Merchant    string            `json:"merchant" bson:"merchant"`
	Location    string            `json:"location" bson:"location"`
	IPAddress   string            `json:"ip_address" bson:"ip_address"`
	Timestamp   time.Time         `json:"timestamp" bson:"timestamp"`
	Status      TransactionStatus `json:"status" bson:"status"`
	FraudResult FraudResult       `json:"fraud_result" bson:"fraud_result"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" bson:"updated_at"`
}
