package main

import (
	"context"
	"database/sql"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"os"
	"shop/order/internal/handler"
	"shop/order/internal/repository"
	"shop/order/internal/service"
	"shop/pkg/broker"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"
)

func main() {
	commandsTopic := "order-commands"

	logger := log.New(os.Stdout, "[order] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("postgres", "postgres://order:order@localhost:5436/order?sslmode=disable")
	if err != nil {
		logger.Fatal("failed to connect to database", "error", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed to ping database", "error", err)
	}

	ctx := context.Background()

	out := outbox.NewPostgresOutbox()

	orderRepo := repository.NewPostgresOrderRepository()
	orderService := service.NewOrderService(orderRepo, logger)

	brokers := []string{"localhost:9093"}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Net.MaxOpenRequests = 1
	config.Producer.Transaction.ID = uuid.New().String()

	// kafka producer
	kafkaProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании продюсера: %v", err)
	}
	defer kafkaProducer.Close()

	// kafka consumer
	group := "order-group"
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании потребителя: %v", err)
	}
	defer kafkaConsumer.Close()

	in := inbox.NewPostgresInbox()
	commandHandler := handler.NewCommandHandler(db, orderService, in, out, logger)
	br := broker.NewKafkaBroker(kafkaProducer, kafkaConsumer, logger)

	// worker
	workerBatchSize := 100
	workerInterval := 1 * time.Second
	outboxWorker := outbox.NewWorker(db, br, out, logger, workerBatchSize, workerInterval)
	go func() {
		err := outboxWorker.Start(ctx)
		if err != nil {
			logger.Printf("failed to start outbox worker: %v", err)
		}
	}()

	// subscribe handler
	err = br.Subscribe(commandsTopic, commandHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	br.StartConsume([]string{commandsTopic})

	select {}
}
