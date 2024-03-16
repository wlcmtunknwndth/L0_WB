package nats_server

import (
	"encoding/json"
	"github.com/nats-io/stan.go"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"log/slog"
	"net/http"
)

type Storage interface {
	SaveOrder(order *storage.Order) error
	GetOrder(uuid string) (*storage.Order, error)
}

type Broker struct {
	sc stan.Conn
	db *Storage
}

func New(cfg *config.Config, db Storage) *Broker {
	sc, err := stan.Connect(
		"test-cluster",
		"db-saver",
		stan.Pings(1, 3),
		stan.NatsURL(cfg.Nats.IpAddr),
	)
	if err != nil {
		slog.Error("couldn't run nats server")
	}

	return &Broker{db: &db, sc: sc}
}

const (
	SendOrder = "getOrder"
	SaveOrder = "saveOrder"
)

func (b *Broker) Saver() (stan.Subscription, error) {
	sub, err := b.sc.Subscribe(SaveOrder, func(m *stan.Msg) {
		var order storage.Order
		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			slog.Error("couldn't unmarshal order: ", err)
			return
		}

		err = (*b.db).SaveOrder(&order) // fix
		if err != nil {
			slog.Error("couldn't save order: ", err)
			return
		}
	})
	if err != nil {
		slog.Error("couldn't run channel: ", err)
		return nil, err
	}
	return sub, nil
}

func (b *Broker) PublishOrder(order []byte) error {
	err := b.sc.Publish(SaveOrder, order)
	if err != nil {
		slog.Error("couldn't publish order to save: ", err)
		return err
	}
	return nil
}

func (b *Broker) PublishUUID(uuid []byte) error {
	err := b.sc.Publish(SendOrder, uuid)
	if err != nil {
		slog.Error("couldn't publish order to save: ", err)
		return err
	}
	return nil
}

func (b *Broker) GetHandler() (stan.Subscription, error) {
	sub, err := b.sc.Subscribe(SendOrder, func(m *stan.Msg) {
		var uuid = string(m.Data)

		order, err := (*b.db).GetOrder(uuid)
		if err != nil {
			slog.Error("couldn't get order from storage: ", err)
			return
		}

		ans, err := json.Marshal(order)
		if err != nil {
			slog.Error("couldn't encode order: ", err)
			return
		}

		if err = b.sc.Publish(order.OrderID, ans); err != nil {
			slog.Error("couldn't publish order: ", err)
			return
		}
	})
	if err != nil {
		slog.Error("couldn't run get handler")
		return nil, err
	}
	return sub, nil
}

func (b *Broker) OrderGetter(uuid string, w http.ResponseWriter) (stan.Subscription, error) {
	sub, err := b.sc.Subscribe(uuid, func(m *stan.Msg) {
		var order storage.Order
		if err := json.Unmarshal(m.Data, &order); err != nil {
			slog.Error("couldn't unmarshal message: ", err)
			return
		}

		if _, err := w.Write(m.Data); err != nil {
			slog.Error("couldn't write respond")
			return
		}
	})
	if err != nil {
		slog.Error("couldn't run order getter: ", err)
		return nil, err
	}

	return sub, nil
}
