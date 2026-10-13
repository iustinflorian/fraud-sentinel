package db

import (
	"context"
	"fmt"
	"fraud-sentinel/pkg/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TransactionRepository struct {
	db *mongo.Database
}

func NewTransactionRepository(db *mongo.Database) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx *models.TransactionRecord) error {
	if tx.ID == "" {
		tx.ID = bson.NewObjectID().Hex()
	}

	coll := r.db.Collection("transactions")

	_, err := coll.InsertOne(ctx, tx)
	if err != nil {
		return fmt.Errorf("can't save transaction to db: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetTransactionByID(ctx context.Context, id string) (*models.TransactionRecord, error) {
	filter := bson.M{"_id": id}

	var tx models.TransactionRecord

	err := r.db.Collection("transactions").FindOne(ctx, filter).Decode(&tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (r *TransactionRepository) UpdateFraudAssessment(ctx context.Context, id string, status models.TransactionStatus, result models.FraudResult) error {
	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"fraud_result": result,
			"updated_at":   time.Now().UTC(),
		},
	}

	_, err := r.db.Collection("transactions").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("can't update transaction %s: %w", id, err)
	}

	return nil
}
