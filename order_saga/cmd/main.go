package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"shop/order_saga/internal/handler"
	"shop/order_saga/internal/orchestrator"
	"shop/order_saga/internal/repository"
	"shop/order_saga/internal/service"
	"shop/pkg/broker"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	commandsTopic := "order-saga-commands"
	productEventTopic := "product-events"
	inventoryEventTopic := "inventory-events"
	orderEventTopic := "order-events"
	paymentEventTopic := "payment-events"

	logger := log.New(os.Stdout, "[order-saga] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("pgx", "postgres://saga:saga@localhost:5437/saga?sslmode=disable")
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

	orderRepo := repository.NewPostgresSagaRepo()
	orc := orchestrator.NewOrchestrator(orderRepo, out, logger)
	orderSagaService := service.NewOrderSagaService(orc, logger)

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
	group := "order-saga-group"
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании потребителя: %v", err)
	}
	defer kafkaConsumer.Close()

	in := inbox.NewPostgresInbox()
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

	//subscribe command handler
	commandHandler := handler.NewCommandHandler(db, orderSagaService, in, out, logger)
	err = br.Subscribe(commandsTopic, commandHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	// subscribe event handler
	eventHandler := handler.NewEventHandler(db, orc, in, out, logger)
	err = br.Subscribe(productEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}
	err = br.Subscribe(inventoryEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}
	err = br.Subscribe(orderEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}
	err = br.Subscribe(paymentEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	br.StartConsume([]string{
		commandsTopic,
		productEventTopic,
		inventoryEventTopic,
		orderEventTopic,
		paymentEventTopic,
	})

	logger.Println("wwwwwwwwwwww")

	select {}
}
