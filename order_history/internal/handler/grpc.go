package handler

import (
	"context"
	"database/sql"
	"log"
	"shop/order_history/internal/repository"
	"shop/order_history/proto"
)

type GrpcHandler struct {
	proto.UnimplementedOrderHistoryServiceServer
	db        *sql.DB
	orderRepo repository.OrderRepository
	logger    *log.Logger
}

func NewGrpcHandler(db *sql.DB, orderRepo repository.OrderRepository, logger *log.Logger) *GrpcHandler {
	return &GrpcHandler{db: db, orderRepo: orderRepo, logger: logger}
}

func (h *GrpcHandler) GetOrders(ctx context.Context, in *proto.GetOrdersRequest) (*proto.GetOrdersResponse, error) {
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Println("Failed to begin transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(ctx, "tx", tx)

	orders, err := h.orderRepo.GetByUserID(ctxWithTx, in.GetUserId(), int(in.GetPage()), int(in.GetLimit()))
	if err != nil {
		h.logger.Printf("Failed to get orders: %+v", err)
		return nil, err
	}

	var protoOrders []*proto.Order
	for _, order := range orders {
		var protoOrderItems []*proto.OrderItem
		for _, orderItem := range order.OrderItems {
			protoOrderItems = append(protoOrderItems, &proto.OrderItem{
				ProductId: orderItem.ProductID,
				Name:      orderItem.Name,
				Price:     int64(orderItem.Price),
				Quantity:  int64(orderItem.Quantity),
			})
		}

		protoOrders = append(protoOrders, &proto.Order{
			Id:                order.ID,
			OrderItems:        protoOrderItems,
			PaymentId:         order.PaymentID,
			PaymentMethodId:   order.PaymentMethodID,
			PaymentType:       order.PaymentType,
			PaymentGateway:    order.PaymentGateway,
			PaymentSum:        int64(order.PaymentSum),
			PaymentExternalId: order.PaymentExternalID,
			PaymentStatus:     order.PaymentStatus,
			Status:            order.PaymentStatus,
			CreatedAt:         order.CreatedAt.String(),
			UpdatedAt:         order.UpdatedAt.String(),
		})
	}

	err = tx.Commit()
	if err != nil {
		h.logger.Println("Failed to commit transaction", "error", err)
		return nil, err
	}

	return &proto.GetOrdersResponse{Orders: protoOrders}, nil
}
