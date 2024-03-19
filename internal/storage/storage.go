package storage

import (
	"github.com/brianvoe/gofakeit/v6"
	"time"
)

type Order struct {
	OrderID           string    `json:"order_uid" protobuf:"order_uid"`
	TrackNum          string    `json:"track_number" protobuf:"track_number"`
	Entry             string    `json:"entry" protobuf:"WBIL"`
	Delivery          Delivery  `json:"delivery" protobuf:"delivery"`
	Payment           Payment   `json:"payment" protobuf:"payment"`
	Items             []Item    `json:"items" protobuf:"items"`
	Locale            string    `json:"locale" protobuf:"locale"`
	InternalSignature string    `json:"internal_signature" protobuf:"internal_signature"`
	CustomerId        string    `json:"customer_id" protobuf:"customer_id"`
	DeliveryService   string    `json:"delivery_service" protobuf:"delivery_service"`
	Shardkey          string    `json:"shardkey" protobuf:"shardkey"`
	SmId              uint32    `json:"sm_id" protobuf:"sm_id"`
	DateCreated       time.Time `json:"date_created" protobuf:"date_created"`
	OofShard          string    `json:"oof_shard" protobuf:"oof_shard"`
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
	Amount       uint32 `json:"amount"`
	PaymentDt    uint32 `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost uint16 `json:"delivery_cost"`
	GoodsTotal   uint32 `json:"goods_total"`
	CustomFee    uint16 `json:"custom_fee"`
}

type Item struct {
	ChrtID      uint32 `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       uint32 `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        uint8  `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  uint32 `json:"total_price"`
	NmID        uint32 `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      uint8  `json:"status"`
}

// SearchRequest -- needed for unmarshaling to search for the storage.Order in the backend by sent uuid.
type SearchRequest struct {
	Uuid string `json:"order_uid"`
}

var banks = []string{"sber", "alpha", "raif", "tinkoff", "wtb", "rnkb"}
var currency = []string{"USD", "RUB", "GBP", "EUR", "GRN", "IDK"}

// RandomOrder -- creates order with random fields.
func RandomOrder(uuid string) *Order {
	address := gofakeit.Address()
	//uuid := gofakeit.UUID()
	trackNum := gofakeit.HexUint64()

	return &Order{
		OrderID:  uuid,
		TrackNum: trackNum,
		Entry:    "WBIL",
		Delivery: Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     address.Zip,
			City:    address.City,
			Address: address.Street,
			Region:  address.State,
			Email:   gofakeit.Email(),
		},
		Payment: Payment{
			Transaction:  uuid,
			ReqID:        "",
			Currency:     gofakeit.RandomString(currency),
			Provider:     "wbpay",
			Amount:       gofakeit.Uint32(),
			PaymentDt:    gofakeit.Uint32(),
			Bank:         gofakeit.RandomString(banks),
			DeliveryCost: gofakeit.Uint16(),
			GoodsTotal:   gofakeit.Uint32(),
			CustomFee:    gofakeit.Uint16(),
		},
		Items: []Item{
			Item{
				ChrtID:      gofakeit.Uint32(),
				TrackNumber: trackNum,
				Price:       uint32(gofakeit.IntRange(1, 10000000)),
				Rid:         gofakeit.UUID(),
				Name:        gofakeit.Name(),
				Sale:        uint8(gofakeit.IntRange(1, 100)),
				Size:        gofakeit.HexUint8(),
				TotalPrice:  uint32(gofakeit.IntRange(1, 10000000)),
				NmID:        uint32(gofakeit.IntRange(1, 100000)),
				Brand:       gofakeit.Name(),
				Status:      uint8(gofakeit.HTTPStatusCode()),
			},
			Item{
				ChrtID:      gofakeit.Uint32(),
				TrackNumber: trackNum,
				Price:       uint32(gofakeit.IntRange(1, 10000000)),
				Rid:         gofakeit.UUID(),
				Name:        gofakeit.Name(),
				Sale:        uint8(gofakeit.IntRange(1, 100)),
				Size:        gofakeit.HexUint8(),
				TotalPrice:  uint32(gofakeit.IntRange(1, 10000000)),
				NmID:        uint32(gofakeit.IntRange(1, 100000)),
				Brand:       gofakeit.Name(),
				Status:      uint8(gofakeit.HTTPStatusCode()),
			},
		},
		Locale:            gofakeit.Language(),
		InternalSignature: "",
		CustomerId:        gofakeit.UUID(),
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmId:              gofakeit.Uint32(),
		DateCreated:       gofakeit.Date(),
		OofShard:          "1",
	}
}
