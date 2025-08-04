package domain

/*
	 type Order struct {
		OrderUID          string  `json:"order_uid"`
		Entry             string  `json:"entry"`
		InternalSignature string  `json:"internal_signature"`
		Payment           Payment `json:"payment"`
		Items             []Items `json:"items"`
		Locale            string  `json:"locale"`
		CustomerID        string  `json:"customer_id"`
		TrackNumber       string  `json:"track_number"`
		DeliveryService   string  `json:"delivery_service"`
		Shardkey          string  `json:"shardkey"`
		SmID              int     `json:"sm_id"`
	}

	type OrderOut struct {
		OrderUID        string `json:"order_uid"`
		Entry           string `json:"entry"`
		TotalPrice      int    `json:"total_price"`
		CustomerID      string `json:"customer_id"`
		TrackNumber     string `json:"track_number"`
		DeliveryService string `json:"delivery_service"`
	}
*/
type Order struct {
	OrderUID          string   `json:"order_uid"`
	Entry             string   `json:"entry" validate:"required"`
	InternalSignature string   `json:"internal_signature"`
	Payment           Payment  `json:"payment" validate:"required"`
	Items             []Items  `json:"items" validate:"required,dive,required"` // Dive into the slice and validate each item
	Locale            string   `json:"locale" validate:"required"`
	CustomerID        string   `json:"customer_id" validate:"required"`
	TrackNumber       string   `json:"track_number" validate:"required"`
	DeliveryService   string   `json:"delivery_service" validate:"required"`
	Shardkey          string   `json:"shardkey" validate:"required"`
	SmID              int      `json:"sm_id" validate:"required"`
	Delivery          Delivery `json:"delivery" validate:"required"` // Add validation for Delivery struct
}

type OrderOut struct {
	OrderUID        string `json:"order_uid"`
	Entry           string `json:"entry"`
	TotalPrice      int    `json:"total_price"`
	CustomerID      string `json:"customer_id"`
	TrackNumber     string `json:"track_number"`
	DeliveryService string `json:"delivery_service"`
}
