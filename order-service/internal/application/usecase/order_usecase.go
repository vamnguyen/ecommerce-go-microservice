package usecase

import (
	"context"

	"order-service/internal/application/dto"
	"order-service/internal/domain/entity"
	domainErr "order-service/internal/domain/errors"
	"order-service/internal/domain/repository"
	"order-service/internal/infrastructure/client"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	orderRepo repository.OrderRepository
	userClient *client.UserClient
}

func NewOrderUseCase(orderRepo repository.OrderRepository, userClient *client.UserClient) *OrderUseCase {
	return &OrderUseCase{
		orderRepo: orderRepo,
		userClient: userClient,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*dto.OrderDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	// Validate user-service is available
	err = uc.userClient.GetUser(ctx, userID)
	if err != nil {
		// For MVP, we'll continue even if user-service is unavailable
		// In production, you might want to fail or use a circuit breaker
	}

	if len(req.Items) == 0 {
		return nil, domainErr.ErrInvalidInput
	}

	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		if item.Quantity <= 0 || item.Price < 0 {
			return nil, domainErr.ErrInvalidInput
		}
		items[i] = entity.NewOrderItem(
			uuid.Nil, // Will be set after order creation
			item.ProductID,
			item.ProductName,
			item.Quantity,
			item.Price,
		)
	}

	order := entity.NewOrder(
		userUUID,
		items,
		req.ShippingAddress,
		req.ShippingCity,
		req.ShippingCountry,
		req.ShippingPostalCode,
	)

	// Set order ID for items
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	return uc.toDTO(order), nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, orderID string, userID string) (*dto.OrderDTO, error) {
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	order, err := uc.orderRepo.FindByID(ctx, orderUUID)
	if err != nil {
		return nil, err
	}

	// Users can only view their own orders (unless admin)
	userUUID, _ := uuid.Parse(userID)
	if order.UserID != userUUID {
		return nil, domainErr.ErrForbidden
	}

	return uc.toDTO(order), nil
}

func (uc *OrderUseCase) ListOrders(ctx context.Context, userID string, page, pageSize int32) (*dto.ListOrdersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	limit := int(pageSize)
	offset := int((page - 1) * pageSize)

	orders, total, err := uc.orderRepo.FindByUserID(ctx, userUUID, limit, offset)
	if err != nil {
		return nil, err
	}

	orderInfos := make([]dto.OrderInfoDTO, len(orders))
	for i, order := range orders {
		orderInfos[i] = dto.OrderInfoDTO{
			OrderID:     order.ID.String(),
			UserID:      order.UserID.String(),
			Status:      string(order.Status),
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &dto.ListOrdersResponse{
		Orders:   orderInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, orderID string, userID string, status string) error {
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	order, err := uc.orderRepo.FindByID(ctx, orderUUID)
	if err != nil {
		return err
	}

	// Users can only update their own orders (unless admin)
	userUUID, _ := uuid.Parse(userID)
	if order.UserID != userUUID {
		return domainErr.ErrForbidden
	}

	newStatus := entity.OrderStatus(status)
	if !order.CanUpdateStatus(newStatus) {
		return domainErr.ErrStatusTransition
	}

	order.UpdateStatus(newStatus)

	return uc.orderRepo.Update(ctx, order)
}

func (uc *OrderUseCase) toDTO(order *entity.Order) *dto.OrderDTO {
	items := make([]dto.OrderItemDTO, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemDTO{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		}
	}

	return &dto.OrderDTO{
		OrderID:           order.ID.String(),
		UserID:            order.UserID.String(),
		Status:            string(order.Status),
		TotalAmount:       order.TotalAmount,
		Items:             items,
		ShippingAddress:   order.ShippingAddress,
		ShippingCity:      order.ShippingCity,
		ShippingCountry:   order.ShippingCountry,
		ShippingPostalCode: order.ShippingPostalCode,
		CreatedAt:         order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

