package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Status            OrderStatus
	TotalAmount       float64
	Items             []OrderItem
	ShippingAddress   string
	ShippingCity      string
	ShippingCountry   string
	ShippingPostalCode string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   string
	ProductName string
	Quantity    int32
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewOrder(userID uuid.UUID, items []OrderItem, shippingAddress, shippingCity, shippingCountry, shippingPostalCode string) *Order {
	now := time.Now()
	
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	return &Order{
		ID:                uuid.New(),
		UserID:            userID,
		Status:            OrderStatusPending,
		TotalAmount:       totalAmount,
		Items:             items,
		ShippingAddress:   shippingAddress,
		ShippingCity:      shippingCity,
		ShippingCountry:   shippingCountry,
		ShippingPostalCode: shippingPostalCode,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func NewOrderItem(orderID uuid.UUID, productID, productName string, quantity int32, price float64) OrderItem {
	now := time.Now()
	return OrderItem{
		ID:          uuid.New(),
		OrderID:     orderID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

func (o *Order) CanUpdateStatus(newStatus OrderStatus) bool {
	switch o.Status {
	case OrderStatusPending:
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		return newStatus == OrderStatusShipped || newStatus == OrderStatusCancelled
	case OrderStatusShipped:
		return newStatus == OrderStatusDelivered
	case OrderStatusDelivered, OrderStatusCancelled:
		return false
	default:
		return false
	}
}

