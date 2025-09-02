package inbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

type PostgresInbox struct{}

func NewPostgresInbox() *PostgresInbox {
	return &PostgresInbox{}
}

func (o *PostgresInbox) Store(ctx context.Context, message Message) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}
	jsonPayload, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	query := "INSERT INTO inbox (message_id, message_type, topic, key, payload, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err = tx.ExecContext(ctx, query, message.MessageID, message.MessageType, message.Topic, message.Key, jsonPayload, message.Status, message.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (o *PostgresInbox) Exists(ctx context.Context, messageID string) (bool, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return false, errors.New("transaction not found in context")
	}

	query := `SELECT 1 from inbox WHERE message_id = $1`
	var result int
	err := tx.QueryRowContext(ctx, query, messageID).Scan(
		&result,
	)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (o *PostgresInbox) MarkAsPending(ctx context.Context, messageID string) error {
	return o.updateStatus(ctx, messageID, StatusInit, StatusPending)
}

func (o *PostgresInbox) MarkAsCompleted(ctx context.Context, messageID string) error {
	return o.updateStatus(ctx, messageID, StatusPending, StatusCompleted)
}

func (o *PostgresInbox) MarkAsError(ctx context.Context, messageID string) error {
	return o.updateStatus(ctx, messageID, StatusPending, StatusError)
}

func (o *PostgresInbox) updateStatus(ctx context.Context, messageID string, from MessageStatus, to MessageStatus) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}

	query := "UPDATE inbox SET status = $1 WHERE message_id = $2 AND status = $3"
	r, err := tx.ExecContext(ctx, query, to, messageID, from)
	if err != nil {
		log.Println("failed to update status value", "error", err)
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected", "error", err)
		return err
	}
	if n == 0 {
		log.Println("failed to update status: no rows affected")
		return errors.New("failed to update status: no rows affected")
	}
	return nil
}
