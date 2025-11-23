package cache

import (
	"sync"
	"wb-l0/internal/model"
	"errors"
)

type Cache struct {
	sync.RWMutex
	orders map[string]model.Order
}

func New() *Cache {
	return &Cache{
		orders: make(map[string]model.Order),
	}
}

func (c *Cache) Set(order model.Order) {
	c.Lock()
	defer c.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(id string) (model.Order, error) {
	c.RLock()
	defer c.RUnlock()
	order, ok := c.orders[id]
	if !ok {
		return model.Order{}, errors.New("order not found")
	}
	return order, nil
}
