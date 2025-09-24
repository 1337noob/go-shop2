package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"shop/payment/internal/handler"
	"shop/payment/internal/repository"
	"shop/payment/internal/service"
	"shop/pkg/broker"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	commandsTopic := "payment-commands"

	logger := log.New(os.Stdout, "[payment] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("postgres", "postgres://payment:payment@localhost:5435/payment?sslmode=disable")
	if err != nil {
		logger.Fatal("failed to connect to database", "error", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed to ping database", "error", err)
	}

	ctx := context.Background()

	o := outbox.NewPostgresOutbox()

	methodRepo := repository.NewPostgresMethodRepository()
	methodService := service.NewMethodService(methodRepo, logger)

	paymentRepo := repository.NewPostgresPaymentRepository()
	paymentService := service.NewPaymentService(paymentRepo, methodRepo, logger)

	brokers := []string{"localhost:9093"}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	// нужен?
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
	group := "payment-group"
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании потребителя: %v", err)
	}
	defer kafkaConsumer.Close()
	in := inbox.NewPostgresInbox()

	commandHandler := handler.NewCommandHandler(db, methodService, paymentService, in, o, logger)
	br := broker.NewKafkaBroker(kafkaProducer, kafkaConsumer, logger)

	// worker
	workerBatchSize := 100
	workerInterval := 1 * time.Second
	outboxWorker := outbox.NewWorker(db, br, o, logger, workerBatchSize, workerInterval)
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
