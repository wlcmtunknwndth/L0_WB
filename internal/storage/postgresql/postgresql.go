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

	if err = db.Ping(); err != nil {
		slog.Error("couldn't ping db", err)
	} else {
		slog.Info("pinged db successfully")
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}

func (s *Storage) GetOrder(orderUid string) (*storage.Order, error) {
	const op = "storage.postrgesql.getOrder"
	result := s.db.QueryRow(getOrderTemplate, orderUid)

	order, err := ParseOrder(result)
	if err != nil {
		slog.Error(op, "Couldn't get order: ", err)
		return nil, err
	}

	res, err := s.db.Query(getItemsTemplate, order.TrackNum)

	order.Items = *ParseItems(res)
	if err != nil {
		slog.Error(op, "Couldn't get items: ", err)
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
		order.TrackNum, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
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

	for i := 0; i < len(order.Items); i++ {
		_, err = s.db.Exec(saveItems,
			order.Items[i].ChrtID, order.Items[i].TrackNumber, order.Items[i].Price,
			order.Items[i].Rid, order.Items[i].Name, order.Items[i].Sale, order.Items[i].Size,
			order.Items[i].TotalPrice, order.Items[i].NmID, order.Items[i].Brand,
			order.Items[i].Status,
		)
		if err != nil {
			return fmt.Errorf("%s: items: %w", op, err)
		}
	}

	return nil

}

func ParseOrder(row *sql.Row) (*storage.Order, error) {
	var order storage.Order
	err := row.Scan( // Common order Info
		&order.OrderID, &order.TrackNum, &order.Entry,
		&order.Locale, &order.InternalSignature,
		&order.CustomerId, &order.DeliveryService,
		&order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard,
		//Delivery
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
		//Payment
		&order.Payment.ReqID,
		&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	order.Payment.Transaction = order.OrderID

	if err != nil {
		return nil, fmt.Errorf("error parsing order: %w", err)
	}

	return &order, nil
}

func ParseItems(row *sql.Rows) *[]storage.Item {
	var items []storage.Item = make([]storage.Item, 0)

	for row.Next() {
		var item storage.Item
		err := row.Scan(&item.ChrtID, &item.TrackNumber, &item.Price,
			&item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand,
			&item.Status)
		if err != nil {
			slog.Error("Error parsing item:", err)
			return nil
		}
		items = append(items, item)
	}

	return &items
}
