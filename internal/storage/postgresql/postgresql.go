package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"log/slog"
)

type Storage struct {
	db *sql.DB
}

// New -- creates new instance of storage.Storage.
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

// Delete -- deletes storage.Order from storage.
func (s *Storage) Delete(uuid, trackNum string) error {
	_, err := s.db.Exec(deleteItems, trackNum)
	if err != nil {
		slog.Error("couldn't delete order", err)
		return err
	}

	_, err = s.db.Exec(deleteDelivery, trackNum)
	if err != nil {
		slog.Error("couldn't delete order", err)
		return err
	}

	_, err = s.db.Exec(deletePayment, uuid)
	if err != nil {
		slog.Error("couldn't delete order", err)
		return err
	}

	_, err = s.db.Exec(deleteOrder, uuid)
	if err != nil {
		slog.Error("couldn't delete order", err)
		return err
	}

	return nil
}

// GetOrder -- sends storage.Order by given uid if exists.
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

// SaveOrder -- saves the given order to the storage.
func (s *Storage) SaveOrder(order *storage.Order) error {
	const op = "storage.postgresql.SaveOrder"

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

// ParseOrder -- parses sql.Row to storage.Order.
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

// ParseItems -- parses sql.Row from storage to []storage.Item.
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

// DeleteCache -- deletes cached uuid from storage.
func (s *Storage) DeleteCache(uuid string) error {
	_, err := s.db.Exec(deleteCache, uuid)
	if err != nil {
		slog.Error("couldn't delete from cached: ", err)
	}
	return err
}

// SaveCache -- saves cache to the storage.
func (s *Storage) SaveCache(uuid string) error {
	_, err := s.db.Exec(saveCache, uuid)
	if err != nil {
		//slog.Error("couldn't save cache: ", err)
		return err
	}
	return nil
}

// IsAlreadyCached -- checks if uuid cache has already been saved to the storage.
func (s *Storage) IsAlreadyCached(uuid string) bool {
	var uuidRow string
	if err := s.db.QueryRow(isAlreadyCached, uuid).Scan(&uuidRow); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		slog.Error("couldn't check if order is cached")
		return false
	}
	return true
}

// RestoreCache -- returns []storage.Order by backupED uuids in the storage.
func (s *Storage) RestoreCache() (*[]storage.Order, error) {
	rows, err := s.db.Query(getCache)
	if err != nil {
		slog.Error("couldn't query restoring cache: ", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("couldn't close rows")
		}
	}(rows)
	uuids := make([]string, 0)

	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			slog.Error("couldn't get all the uuids from cache: ", err)
			continue
		}
		uuids = append(uuids, tmp)
	}

	orders := make([]storage.Order, 0)
	for _, value := range uuids {
		order, err := s.GetOrder(value)
		if err != nil {
			slog.Error("couldn't get order: ", value, err)
			break
		}
		orders = append(orders, *order)
	}
	return &orders, err
}
