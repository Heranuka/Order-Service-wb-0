package domain

/* type Payment struct {
	Transaction  string `json:"transaction"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
}
*/

type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	Currency     string `json:"currency" validate:"required,len=3"` // Assuming currency is a 3-letter code
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,min=0"`
	PaymentDt    int    `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"required,min=0"`
}
