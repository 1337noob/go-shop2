package handler

import (
	"context"
	"database/sql"
	"log"
	"shop/pkg/proto"
	"shop/product/internal/repository"
)

type GrpcHandler struct {
	proto.UnimplementedProductServiceServer
	db           *sql.DB
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
	logger       *log.Logger
}

func NewGrpcHandler(db *sql.DB, categoryRepo repository.CategoryRepository, productRepo repository.ProductRepository, logger *log.Logger) *GrpcHandler {
	return &GrpcHandler{db: db, categoryRepo: categoryRepo, productRepo: productRepo, logger: logger}
}

func (h *GrpcHandler) GetCategories(ctx context.Context, in *proto.GetCategoriesRequest) (*proto.GetCategoriesResponse, error) {
	//tx, err := h.db.Begin()
	//if err != nil {
	//	h.logger.Println("Failed to begin transaction", "error", err)
	//	return nil, err
	//}
	//defer tx.Rollback()
	//
	//ctxWithTx := context.WithValue(ctx, "tx", tx)

	categories, err := h.categoryRepo.Get(ctx, int(in.GetPage()), int(in.GetLimit()))
	if err != nil {
		h.logger.Printf("Failed to get categories: %+v", err)
		return nil, err
	}

	var protoCategories []*proto.Category
	for _, category := range categories {
		protoCategories = append(protoCategories, &proto.Category{
			Id:        category.ID,
			Name:      category.Name,
			CreatedAt: category.CreatedAt.String(),
		})
	}

	//err = tx.Commit()
	//if err != nil {
	//	h.logger.Println("Failed to commit transaction", "error", err)
	//	return nil, err
	//}

	return &proto.GetCategoriesResponse{Categories: protoCategories}, nil
}

func (h *GrpcHandler) GetProductsByCategoryId(ctx context.Context, in *proto.GetProductsRequest) (*proto.GetProductsResponse, error) {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Println("Failed to begin transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(ctx, "tx", tx)

	products, err := h.productRepo.GetByCategoryId(ctxWithTx, in.CategoryId, int(in.GetPage()), int(in.GetLimit()))
	if err != nil {
		h.logger.Printf("Failed to get products: %+v", err)
		return nil, err
	}

	var protoProducts []*proto.Product
	for _, prod := range products {
		protoProducts = append(protoProducts, &proto.Product{
			Id:         prod.ID,
			Name:       prod.Name,
			Price:      int64(prod.Price),
			CategoryId: prod.CategoryID,
			CreatedAt:  prod.CreatedAt.String(),
		})
	}

	err = tx.Commit()
	if err != nil {
		h.logger.Println("Failed to commit transaction", "error", err)
		return nil, err
	}

	return &proto.GetProductsResponse{Products: protoProducts}, nil
}
