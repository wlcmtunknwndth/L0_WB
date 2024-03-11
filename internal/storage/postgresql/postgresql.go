package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"log/slog"
)

type Storage struct {
	db *sql.DB
}

func New(config config.DbConfig) (*Storage, error) {
	const op = "storage.postgresql.New"
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", config.DbUser, config.DbPass, config.DbName, config.SSLmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}

func (s *Storage) GetOrder(orderUid string) (*storage.Order, error) {
	const op = "storage.postrgesql.getOrder"
	result := s.db.QueryRow(getTemplate, orderUid)

	order, err := ParseOrder(result)
	if err != nil {
		slog.Error(op, "Couldn't get order: ", err)
		return nil, err
	}
	return order, nil
}

func (s *Storage) SaveOrder(order *storage.Order) error {
	const op = "storage.postgresql.CreateOrder"

	_, err := s.db.Exec(saveOrder,
		order.OrderID, order.TrackNum, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerId, order.DeliveryService,
		order.Shardkey, order.SmId, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("%s: order: %w", op, err)
	}

	_, err = s.db.Exec(saveDelivery,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("%s: delivery: %w", op, err)
	}

	_, err = s.db.Exec(savePayment,
		order.Payment.Transaction, order.Payment.ReqID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt,
		order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("%s: payment: %w", op, err)
	}

	_, err = s.db.Exec(saveItems,
		order.Items.ChrtID, order.Items.TrackNumber, order.Items.Price,
		order.Items.Rid, order.Items.Name, order.Items.Sale, order.Items.Size,
		order.Items.TotalPrice, order.Items.NmID, order.Items.Brand,
		order.Items.Status,
	)
	if err != nil {
		return fmt.Errorf("%s: order: %w", op, err)
	}

	return nil

}

func ParseOrder(row *sql.Row) (*storage.Order, error) {
	var order storage.Order
	var emptyStuff byte
	err := row.Scan( // Common order Info
		&order.OrderID, &order.TrackNum, &order.Entry,
		&emptyStuff, &order.Locale, &order.InternalSignature,
		&order.CustomerId, &order.DeliveryService,
		&order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard,
		//Delivery
		&emptyStuff, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
		//Payment
		&order.Payment.Transaction, &order.Payment.ReqID,
		&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
		//Items
		&order.Items.ChrtID, &order.Items.TrackNumber, &order.Items.Price,
		&order.Items.Rid, &order.Items.Name, &order.Items.Sale, &order.Items.Size,
		&order.Items.TotalPrice, &order.Items.NmID, &order.Items.Brand,
		&order.Items.Status,
	)

	if err != nil {
		return nil, err
	}
	return &order, nil
}

/*
DROP TABLE orders, delivery, items, payment;

CREATE TABLE IF NOT EXISTS orders
(
    order_uid  VARCHAR(128) PRIMARY KEY,
    track_number  VARCHAR(128) UNIQUE,
    entry  VARCHAR(64),
    delivery  INT UNIQUE,
    payment  VARCHAR(128) UNIQUE,
    items  INT UNIQUE,
    locale  VARCHAR(10),
    internal_signature  VARCHAR(128),
    customer_id  VARCHAR(128),
    delivery_service  VARCHAR(128),
    shardkey VARCHAR(64),
    sm_id  INT,
    date_created  timestamp,
    oof_shard  VARCHAR(32)
);

CREATE TABLE IF NOT EXISTS delivery
(
	uuid INT PRIMARY KEY,
	fio VARCHAR(64),
	phone VARCHAR(16),
	zip VARCHAR(16),
	city VARCHAR(32),
	address VARCHAR(64),
	region VARCHAR(32),
	email VARCHAR(128),
	FOREIGN KEY (uuid) REFERENCES orders(delivery)
);

CREATE TABLE IF NOT EXISTS payment
(
	transact VARCHAR(128) UNIQUE,
	request_id VARCHAR(128),
	currency VARCHAR(8),
	provider VARCHAR(32),
	amount INT CHECK(delivery_cost > 0),
	payment_dt INT CHECK(payment_dt > 0),
	bank VARCHAR(32),
	delivery_cost INT CHECK(delivery_cost > 0),
	goods_total INT CHECK(goods_total > 0),
	custom_fee INT CHECK(custom_fee > 0),
	FOREIGN KEY (transact) REFERENCES orders(payment)
);

CREATE TABLE IF NOT EXISTS items
(
	chrt_id INT UNIQUE,
	track_number VARCHAR(128) PRIMARY KEY,
	price INT,
	rid VARCHAR(64) UNIQUE,
	iname VARCHAR(32),
	sale INT CHECK (sale > 0),
	isize VARCHAR(16),
	total_price INT CHECK (total_price > 0),
	nm_id INT CHECK (nm_id > 0),
	brand VARCHAR(32),
	status INT CHECK (status > 0),
	FOREIGN KEY (track_number) REFERENCES orders(track_number)
);

*/
