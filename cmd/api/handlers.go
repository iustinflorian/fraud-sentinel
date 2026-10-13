package main

import (
	"encoding/json"
	"fraud-sentinel/pkg/db"
	"fraud-sentinel/pkg/models"
	"fraud-sentinel/pkg/queue"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	repo *db.TransactionRepository
	mq   *queue.RabbitMQ
}

func NewTransactionHandler(repo *db.TransactionRepository, mq *queue.RabbitMQ) *TransactionHandler {
	return &TransactionHandler{
		repo: repo,
		mq:   mq,
	}
}

func (h *TransactionHandler) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var tx models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "JSON invalid", http.StatusBadRequest)
		return
	}

	if tx.AccountID == "" || tx.Amount <= 0 || tx.Currency == "" {
		http.Error(w, "invalid parameters: account_id, amount, and currency are required", http.StatusBadRequest)
		return
	}

	tx.ID = "tx_" + uuid.New().String()
	tx.Timestamp = time.Now().UTC()

	record := models.TransactionRecord{
		ID:        tx.ID,
		AccountID: tx.AccountID,
		Amount:    tx.Amount,
		Currency:  tx.Currency,
		Merchant:  tx.Merchant,
		Location:  tx.Location,
		IPAddress: tx.IPAddress,
		Timestamp: tx.Timestamp,
		Status:    models.StatusPending,
		CreatedAt: tx.Timestamp,
		UpdatedAt: tx.Timestamp,
	}

	if err := h.repo.CreateTransaction(r.Context(), &record); err != nil {
		http.Error(w, "error saving to database", http.StatusInternalServerError)
		return
	}

	if err := h.mq.PublishTransaction(r.Context(), &tx); err != nil {
		http.Error(w, "error sending to message queue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"id":        tx.ID,
		"status":    models.StatusPending,
		"message":   "Transaction queued for fraud analysis",
		"timestamp": tx.Timestamp,
	})
}

func (h *TransactionHandler) GetTransactionStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "'id' parameter is mandatory", http.StatusBadRequest)
		return
	}

	record, err := h.repo.GetTransactionByID(r.Context(), id)
	if err != nil {
		http.Error(w, "no transaction found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}
