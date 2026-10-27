package generics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// Validator defines a contract for models/payloads that can self-validate.
type Validator interface {
	Validate() error
}

// ParseAndValidate reads an HTTP request body, decodes the JSON, and runs self-validation.
func ParseAndValidate[T Validator](w http.ResponseWriter, r *http.Request) (T, error) {
	var payload T

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Invalid JSON payload: %s"}`, err.Error()), http.StatusBadRequest)
		return payload, err
	}

	if err := payload.Validate(); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Validation failed: %s"}`, err.Error()), http.StatusUnprocessableEntity)
		return payload, err
	}

	return payload, nil
}

// BatchSaver defines a contract for any data model that can be batch-inserted via SQL transactions.
type BatchSaver interface {
	GetInsertQuery() string
	GetArgs() []any
}

// SaveBatch handles any data type (T) that satisfies the BatchSaver contract inside a single atomic transaction.
func SaveBatch[T BatchSaver](ctx context.Context, db *sql.DB, items []T) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := items[0].GetInsertQuery()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare batch statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, item.GetArgs()...); err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	return tx.Commit()
}

// Map transforms a slice of type T into a slice of type R using the mapper function.
func Map[T any, R any](items []T, fn func(T) R) []R {
	if items == nil {
		return nil
	}
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// Filter filters elements from a slice using a predicate function.
func Filter[T any](items []T, predicate func(T) bool) []T {
	if items == nil {
		return nil
	}
	var result []T
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Chunk splits a slice into chunks of the given size.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 || len(items) == 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Contains checks if a slice contains a specific comparable element.
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
