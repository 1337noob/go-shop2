package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"shop/order_history/internal/handler"
	"shop/order_history/internal/repository"
	"shop/pkg/broker"
	"shop/pkg/inbox"
	"shop/pkg/proto"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	productEventTopic := "product-events"
	inventoryEventTopic := "inventory-events"
	orderEventTopic := "order-events"
	paymentEventTopic := "payment-events"

	logger := log.New(os.Stdout, "[order_history] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("postgres", "postgres://order_history:order_history@localhost:5439/order_history?sslmode=disable")
	if err != nil {
		logger.Fatal("failed to connect to database", "error", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed to ping database", "error", err)
	}

	orderRepo := repository.NewPostgresOrderRepository()

	brokers := []string{"localhost:9093"}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Net.MaxOpenRequests = 1
	config.Producer.Transaction.ID = uuid.New().String()

	// kafka consumer
	group := "order-history-group"
	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Fatalf("Ошибка при создании потребителя: %v", err)
	}
	defer kafkaConsumer.Close()

	in := inbox.NewPostgresInbox()
	eventHandler := handler.NewEventHandler(db, in, orderRepo, logger)
	br := broker.NewKafkaBroker(nil, kafkaConsumer, logger)

	// subscribe handler
	err = br.Subscribe(orderEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	err = br.Subscribe(productEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	err = br.Subscribe(inventoryEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	err = br.Subscribe(paymentEventTopic, eventHandler)
	if err != nil {
		logger.Fatalf("failed to subscribe to commands topic: %v", err)
	}

	go br.StartConsume([]string{orderEventTopic, productEventTopic, inventoryEventTopic, paymentEventTopic})

	svc := handler.NewGrpcHandler(db, orderRepo, logger)
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}
	logger.Println("Server is listening on :50052")

	srv := grpc.NewServer()
	proto.RegisterOrderHistoryServiceServer(srv, svc)
	logger.Println("gRPC server registered")

	if err = srv.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve: %v", err)
	}

	select {}
}
