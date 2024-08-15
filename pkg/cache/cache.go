package cache

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache struct {
	c *cache.Cache
}

var (
	instance *Cache
	once     sync.Once
)

func GetInstance() *Cache {
	once.Do(func() {
		instance = &Cache{
			c: cache.New(5*time.Minute, 10*time.Minute),
		}
	})
	return instance
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.c.Set(key, value, duration)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.c.Get(key)
}

func (c *Cache) Delete(key string) {
	c.c.Delete(key)
}
