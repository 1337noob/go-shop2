package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"shop/gateway/internal/handler"
	"shop/gateway/internal/middleware"
	"shop/gateway/internal/repository"
	"shop/pkg/broker"
	"shop/pkg/outbox"
	"shop/pkg/proto"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := log.New(os.Stdout, "[gateway] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	db, err := sql.Open("postgres", "postgres://gateway:gateway@localhost:5438/gateway?sslmode=disable")
	if err != nil {
		logger.Fatal("failed to connect to database", "error", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed to ping database", "error", err)
	}
	// Инициализация Redis
	redisRepo, err := repository.NewRedisSessionRepository(
		"localhost:6379",
		"",
		"session:",
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	userRepo := repository.NewPostgresUserRepository(db)

	sessionMiddleware := middleware.NewSessionMiddleware(
		redisRepo,
		"session_id",
		time.Second*60,
	)

	out := outbox.NewPostgresOutbox()

	orderHistoryConn, err := grpc.NewClient(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer orderHistoryConn.Close()
	orderHistoryClient := proto.NewOrderHistoryServiceClient(orderHistoryConn)

	productConn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer productConn.Close()
	productClient := proto.NewProductServiceClient(productConn)

	authHandler := handler.NewAuthHandler(db, sessionMiddleware, userRepo)
	orderHandler := handler.NewOrderHandler(db, out, orderHistoryClient, logger)
	productHandler := handler.NewProductHandler(db, out, productClient, logger)

	router := mux.NewRouter()

	// Public
	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/api/categories", productHandler.GetCategories).Methods("GET")
	router.HandleFunc("/api/products", productHandler.GetProducts).Methods("GET")

	// Protected
	protected := router.PathPrefix("").Subrouter()
	protected.Use(sessionMiddleware.SessionRequired)

	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	protected.HandleFunc("/auth/profile", authHandler.Profile).Methods("GET")

	protected.HandleFunc("/api/orders", orderHandler.CreateOrder).Methods("POST")
	protected.HandleFunc("/api/my-orders", orderHandler.GetMyOrders).Methods("GET")

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

	br := broker.NewKafkaBroker(kafkaProducer, nil, logger)

	// outbox worker
	workerBatchSize := 100
	workerInterval := 1 * time.Second
	outboxWorker := outbox.NewWorker(db, br, out, logger, workerBatchSize, workerInterval)
	go func() {
		err := outboxWorker.Start(context.Background())
		if err != nil {
			logger.Printf("failed to start outbox worker: %v", err)
		}
	}()

	log.Println("API Gateway started on :8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
