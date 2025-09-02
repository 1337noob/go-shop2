package inbox

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Handler interface {
	Handle(ctx context.Context, msg Message) error
}

type Worker struct {
	db        *sql.DB
	inbox     Inbox
	handler   Handler
	logger    *log.Logger
	batchSize int
	interval  time.Duration
}

func NewWorker(db *sql.DB, inbox Inbox, handler Handler, logger *log.Logger, batchSize int, interval time.Duration) *Worker {
	return &Worker{
		db:        db,
		inbox:     inbox,
		handler:   handler,
		logger:    logger,
		batchSize: batchSize,
		interval:  interval,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.logger.Println("starting outbox worker")

	for {
		err := w.process(ctx)
		if err != nil {
			w.logger.Println("failed to process outbox", "error: ", err)
		}
	}
}

func (w *Worker) process(ctx context.Context) error {
	//w.logger.Println("processing outbox")
	//tx, err := w.db.Begin()
	//if err != nil {
	//	w.logger.Fatal("failed to begin transaction", "error", err)
	//}
	//defer tx.Rollback()
	//
	//ctxWithTx := context.WithValue(ctx, "tx", tx)
	//
	//message, err := w.inbox.GetNotCompleted(ctxWithTx)
	//if err != nil {
	//	w.logger.Printf("failed to get messages from outbox: %v", err)
	//	return err
	//}
	//err = w.inbox.MarkAsPending(ctxWithTx, message.ID)
	//if err != nil {
	//	w.logger.Printf("failed to mark messages as pending: %v", err)
	//	return err
	//}
	//
	//err = tx.Commit()
	//if err != nil {
	//	w.logger.Printf("failed to commit transaction", "error", err)
	//	return err
	//}
	//
	//tx, err = w.db.Begin()
	//if err != nil {
	//	w.logger.Fatal("failed to begin transaction", "error", err)
	//}
	//defer tx.Rollback()
	//
	//err = w.handler.Handle(ctxWithTx, message)
	//if err != nil {
	//	w.logger.Printf("failed to handle message: %v", err)
	//	return err
	//}
	//
	//err = w.inbox.MarkAsCompleted(ctxWithTx, message.ID)
	//if err != nil {
	//	w.logger.Printf("failed to mark message as completed: %v", err)
	//	return err
	//}
	//
	//err = tx.Commit()
	//if err != nil {
	//	w.logger.Printf("failed to commit transaction", "error", err)
	//	return err
	//}
	//
	//w.logger.Println("finished outbox")
	return nil

}
