package handler

import (
	"context"
	"log"
	"shop/order_history/internal/repository"
	"shop/pkg/proto"
)

type GrpcHandler struct {
	proto.UnimplementedOrderHistoryServiceServer
	orderRepo repository.OrderRepository
	logger    *log.Logger
}

func NewGrpcHandler(orderRepo repository.OrderRepository, logger *log.Logger) *GrpcHandler {
	return &GrpcHandler{orderRepo: orderRepo, logger: logger}
}

func (h *GrpcHandler) GetOrders(ctx context.Context, in *proto.GetOrdersRequest) (*proto.GetOrdersResponse, error) {
	orders, err := h.orderRepo.GetByUserID(ctx, in.GetUserId(), int(in.GetPage()), int(in.GetLimit()))
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
			Status:            string(order.Status),
			CreatedAt:         order.CreatedAt.String(),
			UpdatedAt:         order.UpdatedAt.String(),
		})
	}

	return &proto.GetOrdersResponse{Orders: protoOrders}, nil
}
