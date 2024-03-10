package storage

type Order struct {
	OrderID           string   `json:"order_uid"`
	TrackNum          string   `json:"track_number"`
	Entry             string   `json:"WBIL"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             Items    `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerId        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmId              string   `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	ReqID        string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       string `json:"amount"`
	PaymentDt    string `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost string `json:"delivery_cost"`
	GoodsTotal   string `json:"goods_total"`
	CustomFee    string `json:"custom_fee"`
}

type Items struct {
	ChrtID      string `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       string `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"Mascaras"`
	Sale        string `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  string `json:"total_price"`
	NmID        string `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      string `json:"status"`
}

//type Items struct {
//	ItemsArr []Item
//}
