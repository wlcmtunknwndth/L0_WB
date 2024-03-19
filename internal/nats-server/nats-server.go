package nats_server

import (
	"encoding/json"
	"github.com/nats-io/stan.go"
	"github.com/wlcmtunknwndth/L0_WB/internal/cacher"
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

// New -- creates a new instance of our Broker, which is needed stan.Conn and storage with methods SaveOrder(order *storage.Order) error and
// GetOrder(uuid string) (*storage.Order, error).
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

// Saver -- saves orders got from streaming channel with the SaveMessage message.
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

// PublishOrder -- publishes order in []byte form with the SaveOrder message, which is listened by Saver.
func (b *Broker) PublishOrder(order []byte) error {
	err := b.sc.Publish(SaveOrder, order)
	if err != nil {
		slog.Error("couldn't publish order to save: ", err)
		return err
	}
	return nil
}

// PublishUUID -- publishes uuid(must be string) in []byte form to the streaming channel with the SendOrder message, so the GetHandler gets the uuid it must
// look for in storage and send back.
func (b *Broker) PublishUUID(uuid []byte) error {
	err := b.sc.Publish(SendOrder, uuid)
	if err != nil {
		slog.Error("couldn't publish order to save: ", err)
		return err
	}
	return nil
}

// GetHandler -- opens subscription to get request. When message is sent, gets the storage.Order from storage.Storage instance with chosen uuid and sends
// it back to streaming channel with uuid of the instance as message, so the other subscription must wait for the message with uuid the user sent.
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

// OrderGetter -- gets order from GetHandler and writes it to our http.ResponseWriter. It func opens subscription by uuid as the name
// and waits for order sent back by GetHandler in []byte, then unmarshals it and write to user response body. Channel are used to verify if data has been sent,
// otherwise the response won't be sent. Returns stan.Subscription to close it later in router(handler).
func (b *Broker) OrderGetter(uuid string, w http.ResponseWriter, ch *chan bool, c *cacher.Cacher) (stan.Subscription, error) {
	sub, err := b.sc.Subscribe(uuid, func(m *stan.Msg) {
		var order storage.Order
		if err := json.Unmarshal(m.Data, &order); err != nil {
			slog.Error("couldn't unmarshal message: ", err)
			return
		}
		c.CacheOrder(order)

		if _, err := w.Write(m.Data); err != nil {
			slog.Error("couldn't write respond")
			return
		}
		*ch <- true
	})
	if err != nil {
		slog.Error("couldn't run order getter: ", err)
		return nil, err
	}

	return sub, nil
}
