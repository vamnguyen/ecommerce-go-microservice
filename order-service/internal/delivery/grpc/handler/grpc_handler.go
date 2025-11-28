package handler

import (
	"context"

	proto "order-service/gen/go"
	"order-service/internal/application/dto"
	"order-service/internal/application/usecase"
	"order-service/internal/delivery/grpc/interceptor"
)

type GRPCHandler struct {
	proto.UnimplementedOrderServiceServer
	orderUsecase usecase.OrderUseCase
}

func NewGRPCHandler(orderUsecase usecase.OrderUseCase) *GRPCHandler {
	return &GRPCHandler{
		orderUsecase: orderUsecase,
	}
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{Status: "healthy", Service: "order service"}, nil
}

func (h *GRPCHandler) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]dto.OrderItemDTO, len(req.GetItems()))
	for i, item := range req.GetItems() {
		items[i] = dto.OrderItemDTO{
			ProductID:   item.GetProductId(),
			ProductName: item.GetProductName(),
			Quantity:    item.GetQuantity(),
			Price:       item.GetPrice(),
		}
	}

	createDTO := dto.CreateOrderRequest{
		Items:             items,
		ShippingAddress:   req.GetShippingAddress(),
		ShippingCity:      req.GetShippingCity(),
		ShippingCountry:   req.GetShippingCountry(),
		ShippingPostalCode: req.GetShippingPostalCode(),
	}

	order, err := h.orderUsecase.CreateOrder(ctx, userID, createDTO)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.CreateOrderResponse{
		OrderId: order.OrderID,
		Message: "order created successfully",
	}, nil
}

func (h *GRPCHandler) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	order, err := h.orderUsecase.GetOrder(ctx, req.GetOrderId(), userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	items := make([]*proto.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &proto.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		}
	}

	return &proto.GetOrderResponse{
		OrderId:           order.OrderID,
		UserId:            order.UserID,
		Status:            order.Status,
		TotalAmount:       order.TotalAmount,
		Items:             items,
		ShippingAddress:   order.ShippingAddress,
		ShippingCity:      order.ShippingCity,
		ShippingCountry:   order.ShippingCountry,
		ShippingPostalCode: order.ShippingPostalCode,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}, nil
}

func (h *GRPCHandler) ListOrders(ctx context.Context, req *proto.ListOrdersRequest) (*proto.ListOrdersResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := h.orderUsecase.ListOrders(ctx, userID, page, pageSize)
	if err != nil {
		return nil, toGRPCError(err)
	}

	orders := make([]*proto.OrderInfo, len(result.Orders))
	for i, order := range result.Orders {
		orders[i] = &proto.OrderInfo{
			OrderId:     order.OrderID,
			UserId:      order.UserID,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt,
		}
	}

	return &proto.ListOrdersResponse{
		Orders:    orders,
		Total:     int32(result.Total),
		Page:      result.Page,
		PageSize:  result.PageSize,
	}, nil
}

func (h *GRPCHandler) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.orderUsecase.UpdateOrderStatus(ctx, req.GetOrderId(), userID, req.GetStatus()); err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.UpdateOrderStatusResponse{Message: "order status updated successfully"}, nil
}

