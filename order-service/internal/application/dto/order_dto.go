package dto

type CreateOrderRequest struct {
	Items            []OrderItemDTO `json:"items"`
	ShippingAddress  string         `json:"shipping_address"`
	ShippingCity     string         `json:"shipping_city"`
	ShippingCountry  string         `json:"shipping_country"`
	ShippingPostalCode string       `json:"shipping_postal_code"`
}

type OrderItemDTO struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int32   `json:"quantity"`
	Price       float64 `json:"price"`
}

type OrderDTO struct {
	OrderID           string         `json:"order_id"`
	UserID            string         `json:"user_id"`
	Status            string         `json:"status"`
	TotalAmount       float64        `json:"total_amount"`
	Items             []OrderItemDTO `json:"items"`
	ShippingAddress   string         `json:"shipping_address"`
	ShippingCity      string         `json:"shipping_city"`
	ShippingCountry   string         `json:"shipping_country"`
	ShippingPostalCode string        `json:"shipping_postal_code"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

type OrderInfoDTO struct {
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt   string  `json:"created_at"`
}

type ListOrdersResponse struct {
	Orders   []OrderInfoDTO `json:"orders"`
	Total    int64          `json:"total"`
	Page     int32          `json:"page"`
	PageSize int32          `json:"page_size"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

