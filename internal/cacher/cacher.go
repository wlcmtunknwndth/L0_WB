package cacher

import (
	"github.com/patrickmn/go-cache"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"log/slog"
	"time"
)

var cached = make(map[string]bool)

type Storage interface {
	RestoreCache() (*[]storage.Order, error)
	SaveCache(uuid string) error
	DeleteCache(uuid string) error
}

type Cacher struct {
	handler *cache.Cache
	db      Storage
}

func New(db Storage, expTime time.Duration, purgeTime time.Duration) *Cacher {
	return &Cacher{
		handler: cache.New(expTime, purgeTime),
		db:      db,
	}
}

func (c *Cacher) CacheOrder(order storage.Order) {
	c.handler.OnEvicted(c.onEvicted)
	c.handler.Set(order.OrderID, order, cache.DefaultExpiration)
	cached[order.OrderID] = true
}

func (c *Cacher) onEvicted(uuid string, data interface{}) {
	delete(cached, uuid)
	err := c.db.DeleteCache(uuid)
	if err != nil {
		slog.Error("couldn't delete order from cache")
	}
}

func (c *Cacher) GetOrder(uuid string) (*storage.Order, bool) {
	data, found := c.handler.Get(uuid)
	if found {
		order := data.(storage.Order)
		return &order, true
	}
	return nil, false
}

func (c *Cacher) Restore() error {
	orders, err := c.db.RestoreCache()
	if err != nil {
		slog.Error("couldn't restore cache: ", err)
		return err
	}

	for i := range *orders {
		c.CacheOrder((*orders)[i])
	}
	return nil
}

func (c *Cacher) SaveCache() error {
	var err error
	for key := range cached {
		err = c.db.SaveCache(key)
		if err != nil {
			slog.Error("couldn't save uuid to cache zone: ", key, err)
			continue
		}
	}
	return nil
}