package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"shop/pkg/outbox"
	"shop/pkg/proto"
	"strconv"
)

type ProductHandler struct {
	db                   *sql.DB
	outbox               outbox.Outbox
	productServiceClient proto.ProductServiceClient
	logger               *log.Logger
}

func NewProductHandler(db *sql.DB, outbox outbox.Outbox, productServiceClient proto.ProductServiceClient, logger *log.Logger) *ProductHandler {
	return &ProductHandler{
		db:                   db,
		outbox:               outbox,
		productServiceClient: productServiceClient,
		logger:               logger,
	}
}

type GetCategoriesRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type GetProductsRequest struct {
	CategoryID string `json:"category_id"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

func (o *ProductHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	//o.logger.Println("GetCategories handler start")

	queryParams := r.URL.Query()

	page := queryParams.Get("page")
	limit := queryParams.Get("limit")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		o.logger.Println("Failed to convert page to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		o.logger.Println("Failed to convert limit to int")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	grpcRequest := proto.GetCategoriesRequest{
		Page:  int64(pageInt),
		Limit: int64(limitInt),
	}

	categories, err := o.productServiceClient.GetCategories(r.Context(), &grpcRequest)
	if err != nil {
		o.logger.Println("Failed to get categories from grpc", "error", err)
		http.Error(w, "Failed to get categories from grpc", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)

	//o.logger.Println("GetCategories handler finish")
}

func (o *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	o.logger.Println("GetProducts handler start")

	var req GetProductsRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	grpcRequest := proto.GetProductsRequest{
		CategoryId: req.CategoryID,
		Page:       int64(req.Page),
		Limit:      int64(req.Limit),
	}

	products, err := o.productServiceClient.GetProductsByCategoryId(r.Context(), &grpcRequest)
	if err != nil {
		o.logger.Println("Failed to get products from grpc", "error", err)
		http.Error(w, "Failed to get products from grpc", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)

	o.logger.Println("GetProducts handler finish")
}
