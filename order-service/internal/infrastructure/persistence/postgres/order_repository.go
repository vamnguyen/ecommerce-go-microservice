package postgres

import (
	"context"
	"errors"
	"time"

	"order-service/internal/domain/entity"
	domainErr "order-service/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderModel struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;index"`
	Status            string    `gorm:"not null;default:'pending'"`
	TotalAmount       float64   `gorm:"not null"`
	ShippingAddress   string
	ShippingCity      string
	ShippingCountry   string
	ShippingPostalCode string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Items             []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string {
	return "orders"
}

type OrderItemModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID   string    `gorm:"not null"`
	ProductName string    `gorm:"not null"`
	Quantity    int32     `gorm:"not null"`
	Price       float64   `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (OrderItemModel) TableName() string {
	return "order_items"
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *entity.Order) error {
	model := r.toModel(order)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	var model OrderModel
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrOrderNotFound
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Order, int64, error) {
	var models []OrderModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&OrderModel{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	orders := make([]*entity.Order, len(models))
	for i, model := range models {
		orders[i] = r.toEntity(&model)
	}

	return orders, total, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *entity.Order) error {
	model := r.toModel(order)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *OrderRepository) List(ctx context.Context, limit, offset int) ([]*entity.Order, int64, error) {
	var models []OrderModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&OrderModel{}).Count(&total).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	if err := r.db.WithContext(ctx).
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	orders := make([]*entity.Order, len(models))
	for i, model := range models {
		orders[i] = r.toEntity(&model)
	}

	return orders, total, nil
}

func (r *OrderRepository) toModel(order *entity.Order) *OrderModel {
	items := make([]OrderItemModel, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemModel{
			ID:          item.ID,
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	return &OrderModel{
		ID:                order.ID,
		UserID:            order.UserID,
		Status:            string(order.Status),
		TotalAmount:       order.TotalAmount,
		ShippingAddress:   order.ShippingAddress,
		ShippingCity:      order.ShippingCity,
		ShippingCountry:   order.ShippingCountry,
		ShippingPostalCode: order.ShippingPostalCode,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
		Items:             items,
	}
}

func (r *OrderRepository) toEntity(model *OrderModel) *entity.Order {
	items := make([]entity.OrderItem, len(model.Items))
	for i, item := range model.Items {
		items[i] = entity.OrderItem{
			ID:          item.ID,
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	return &entity.Order{
		ID:                model.ID,
		UserID:            model.UserID,
		Status:            entity.OrderStatus(model.Status),
		TotalAmount:       model.TotalAmount,
		Items:             items,
		ShippingAddress:   model.ShippingAddress,
		ShippingCity:      model.ShippingCity,
		ShippingCountry:   model.ShippingCountry,
		ShippingPostalCode: model.ShippingPostalCode,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

