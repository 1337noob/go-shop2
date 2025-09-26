package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"shop/pkg/broker"
	"shop/pkg/inbox"
	"shop/pkg/outbox"
	"shop/pkg/proto"
	"shop/product/internal/handler"
	"shop/product/internal/repository"
	"shop/product/internal/service"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
)

func main() {
	commandsTopic := "product-commands"

	logger := log.New(os.Stdout, "[product] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("pgx", "postgres://product:product@localhost:5433/product?sslmode=disable")
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

	catRepo := repository.NewPostgresCategoryRepository(db)
	catService := service.NewCategoryService(catRepo, logger)

	prodRepo := repository.NewPostgresProductRepository()
	prodService := service.NewProductService(prodRepo, logger)

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
	group := "product-group"
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании потребителя: %v", err)
	}
	defer kafkaConsumer.Close()

	in := inbox.NewPostgresInbox()
	commandHandler := handler.NewCommandHandler(db, catService, prodService, in, o, logger)
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

	go br.StartConsume([]string{commandsTopic})

	svc := handler.NewGrpcHandler(db, catRepo, prodRepo, logger)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}
	logger.Println("Server is listening on :50051")

	srv := grpc.NewServer()
	proto.RegisterProductServiceServer(srv, svc)
	logger.Println("gRPC server registered")

	if err = srv.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve: %v", err)
	}

	select {}
}
