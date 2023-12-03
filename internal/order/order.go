package order

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required,max=50"`
	TrackNumber       string    `json:"track_number" validate:"required,max=50"`
	Entry             string    `json:"entry" validate:"required,max=5"`
	Locale            string    `json:"locale" validate:"required,max10"`
	InternalSignature string    `json:"internal_signature"`
	CustomerId        string    `json:"customer_id" validate:"required,max=50"`
	DeliveryService   string    `json:"delivery_service" validate:"required,max=50"`
	ShardKey          string    `json:"shardkey" validate:"required,max=50"`
	SmId              int       `json:"sm_id" validate:"required"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"required,max=10"`
	Delivery          Delivery  `json:"delivery" validate:"required"`
	Payment           Payment   `json:"payment" validate:"required"`
	Items             []Item    `json:"items" validate:"required"`
}

type Delivery struct {
	Name    string `json:"name" validate:"required,max=100"`
	Phone   string `json:"phone" validate:"required,min=12,max=14"`
	Zip     string `json:"zip" validate:"required,max=20"`
	City    string `json:"city" validate:"required,max=50"`
	Address string `json:"address" validate:"required,max=100"`
	Region  string `json:"region" validate:"required,max=100"`
	Email   string `json:"email" validate:"required,min=5,max=345"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required,max=50"`
	RequestId    string `json:"request_id" validate:"max=50"`
	Currency     string `json:"currency" validate:"required,max=3"`
	Provider     string `json:"provider" validate:"required,max=50"`
	Amount       int    `json:"amount" validate:"required,gte=0"`
	PaymentDt    int    `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required,max=50"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"required,gt=0"`
	CustomFee    int    `json:"custom_fee" validate:"required,gte=0"`
}

type Item struct {
	ChrtId      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required,max=50"`
	Price       int    `json:"price" validate:"required,gte=0"`
	Rid         string `json:"rid" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=50"`
	Sale        int    `json:"sale" validate:"required,gte=0"`
	Size        string `json:"size" validate:"required,max=50"`
	TotalPrice  int    `json:"total_price" validate:"required,gte=0"`
	NmId        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required,max50"`
	Status      int    `json:"status" validate:"required"`
}
