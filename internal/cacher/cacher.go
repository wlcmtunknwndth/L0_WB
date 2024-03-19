package cacher

import (
	"github.com/patrickmn/go-cache"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"log/slog"
	"time"
)

// cached -- is the map with saved uuids in current run, so it is easier to back up
var cached = make(map[string]struct{})

type Storage interface {
	RestoreCache() (*[]storage.Order, error)
	SaveCache(uuid string) error
	DeleteCache(uuid string) error
	IsAlreadyCached(uuid string) bool
}

type Cacher struct {
	handler *cache.Cache
	db      Storage
}

// New -- creates new instance of Cacher with Storage interface and cache.Cache vars. expTime -- is the standard expiration time of cached item.
// purgeTime -- is the time the cacher cleans up itself
func New(db Storage, expTime time.Duration, purgeTime time.Duration) *Cacher {
	return &Cacher{
		handler: cache.New(expTime, purgeTime),
		db:      db,
	}
}

// CacheOrder -- caches the order given as an arg and maps order's uuid to cache map.
func (c *Cacher) CacheOrder(order storage.Order) {
	c.handler.OnEvicted(c.onEvicted)
	c.handler.Set(order.OrderID, order, cache.DefaultExpiration)
	//err := c.db.SaveCache(order.OrderID)
	//if err != nil {
	//	slog.Error("couldn't save backup: ", order.OrderID, err)
	//}
	cached[order.OrderID] = struct{}{}
}

// onEvicted -- is a custom func, handling cached item after expiration. It deletes item from cache map and deletes uuid from storage Cache backup.
func (c *Cacher) onEvicted(uuid string, data interface{}) {
	delete(cached, uuid)
	err := c.db.DeleteCache(uuid)
	if err != nil {
		slog.Error("couldn't delete order from cache")
	}
}

// GetOrder -- gets order from cache if found
func (c *Cacher) GetOrder(uuid string) (*storage.Order, bool) {
	data, found := c.handler.Get(uuid)
	if found {
		order := data.(storage.Order)
		return &order, true
	}
	return nil, false
}

// Restore -- restores cached item from backup copy in storage. Must be used at the start of ur application.
func (c *Cacher) Restore() error {
	orders, err := c.db.RestoreCache()
	//fmt.Println(orders)
	if err != nil {
		slog.Error("couldn't restore cache: ", err)
		return err
	}

	for i := range *orders {
		c.CacheOrder((*orders)[i])
	}
	return nil
}

// SaveCache -- backups cache to the storage
func (c *Cacher) SaveCache() error {
	var err error
	for key := range cached {
		if c.db.IsAlreadyCached(key) {
			continue
		}
		err = c.db.SaveCache(key)
		if err != nil {
			slog.Error("couldn't save uuid to cache zone: ", key, err)
			continue
		}
	}
	return nil
}
