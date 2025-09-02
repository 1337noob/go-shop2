package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"log"
)

type PostgresOutbox struct{}

func NewPostgresOutbox() *PostgresOutbox {
	return &PostgresOutbox{}
}

func (o *PostgresOutbox) Publish(ctx context.Context, message Message) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}
	jsonPayload, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	query := "INSERT INTO outbox (id, topic, key, payload, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err = tx.ExecContext(ctx, query, message.ID, message.Topic, message.Key, jsonPayload, message.Status, message.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (o *PostgresOutbox) GetNotSent(ctx context.Context, limit int) ([]Message, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	var messages []Message
	query := "SELECT id, topic, key, payload, status, created_at FROM outbox WHERE status = $1 ORDER BY created_at ASC LIMIT $2"
	rows, err := tx.QueryContext(ctx, query, StatusInit, limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message Message
		var jsonPayload []byte
		err = rows.Scan(&message.ID, &message.Topic, &message.Key, &jsonPayload, &message.Status, &message.CreatedAt)
		if err != nil {
			log.Println("failed to scan row", "error", err)
			return nil, err
		}
		err = json.Unmarshal(jsonPayload, &message.Payload)
		if err != nil {
			log.Println("failed to unmarshal payload", "error", err)
			return nil, err
		}
		messages = append(messages, message)

	}
	//if len(messages) == 0 {
	//	return nil, errors.New("no messages found")
	//}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (o *PostgresOutbox) BatchMarkAsPending(ctx context.Context, ids []string) error {
	return o.batchUpdateStatus(ctx, ids, StatusInit, StatusPending)
}

func (o *PostgresOutbox) BatchMarkAsSent(ctx context.Context, ids []string) error {
	return o.batchUpdateStatus(ctx, ids, StatusPending, StatusSent)
}

func (o *PostgresOutbox) BatchMarkAsError(ctx context.Context, ids []string) error {
	return o.batchUpdateStatus(ctx, ids, StatusPending, StatusError)
}

func (o *PostgresOutbox) batchUpdateStatus(ctx context.Context, ids []string, from MessageStatus, to MessageStatus) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return errors.New("transaction not found in context")
	}
	query := "UPDATE outbox SET status = $1 WHERE id = ANY($2) AND status = $3"
	r, err := tx.ExecContext(ctx, query, to, pq.Array(ids), from)
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
