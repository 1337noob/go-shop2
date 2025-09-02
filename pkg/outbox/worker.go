package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"shop/pkg/broker"
	"time"
)

type Worker struct {
	db        *sql.DB
	broker    broker.Broker
	outbox    Outbox
	logger    *log.Logger
	batchSize int
	interval  time.Duration
}

func NewWorker(db *sql.DB, broker broker.Broker, outbox Outbox, logger *log.Logger, batchSize int, interval time.Duration) *Worker {
	return &Worker{
		db:        db,
		broker:    broker,
		outbox:    outbox,
		logger:    logger,
		batchSize: batchSize,
		interval:  interval,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.logger.Println("starting outbox worker")

	for {
		//time.Sleep(w.interval)
		empty, err := w.processOutbox(ctx)
		if empty {
			w.logger.Printf("outbox is empty, sleep %f seconds", w.interval.Seconds())
			time.Sleep(w.interval)
		}
		if err != nil {
			w.logger.Println("failed to process outbox", "error: ", err)
		}
	}
}

func (w *Worker) processOutbox(ctx context.Context) (bool, error) {
	//w.logger.Println("processing outbox")
	tx, err := w.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		w.logger.Fatal("failed to begin transaction", "error", err)
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(ctx, "tx", tx)

	messages, err := w.outbox.GetNotSent(ctxWithTx, w.batchSize)
	if err != nil {
		w.logger.Printf("failed to get messages from outbox: %v", err)
		return false, err
	}

	if len(messages) == 0 {
		return true, nil
	}

	w.logger.Printf("found %d messages", len(messages))

	var messageIds []string
	for _, m := range messages {
		messageIds = append(messageIds, m.ID)
	}

	err = w.outbox.BatchMarkAsPending(ctxWithTx, messageIds)
	if err != nil {
		w.logger.Printf("failed to mark messages as pending: %v", err)
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		w.logger.Println("failed to commit transaction", "error", err)
		return false, err
	}

	var brokerMessages []broker.Message
	for _, m := range messages {
		jsonPayload, err := json.Marshal(m.Payload)
		if err != nil {
			return false, err
		}
		brokerMessages = append(brokerMessages, broker.Message{
			Topic: m.Topic,
			Key:   m.Key,
			Value: jsonPayload,
		})
	}

	tx, err = w.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		w.logger.Fatal("failed to begin transaction", "error", err)
	}
	defer tx.Rollback()

	ctxWithTx = context.WithValue(ctx, "tx", tx)

	w.logger.Printf("sending messages")

	// send to broker
	err = w.broker.PublishBatch(brokerMessages)
	if err != nil {
		return false, err
	}
	w.logger.Printf("messages sent")

	// mark as sent
	w.logger.Printf("marking messages as sent")
	err = w.outbox.BatchMarkAsSent(ctxWithTx, messageIds)
	if err != nil {
		return false, err
	}
	w.logger.Printf("messages marked as sent")

	err = tx.Commit()
	if err != nil {
		w.logger.Printf("failed to commit transaction", "error", err)
		return false, err
	}

	w.logger.Println("finished outbox")

	return false, nil
}
