package redis

import (
	"log"
	"time"

	c "github.com/asiainfoLDP/datafoundry_oauth2/util/cache"
	"github.com/garyburd/redigo/redis"
)

var (
	REDIS_POOL *redis.Pool
)

func createPool(server, auth string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   10,
		Wait:        true,
		IdleTimeout: 4 * time.Minute,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if auth != "" {
				if _, err := c.Do("AUTH", auth); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

type cache struct {
	pool *redis.Pool
}

func (p *cache) HCache(key, field interface{}, value []byte) error {
	go func() {
		c := p.pool.Get()
		defer c.Close()

		if _, err := c.Do("HSET", key, field, value); err != nil {
			log.Println("[HMap Set] err :", err)
			return
		}

	}()
	return nil
}

func (p *cache) HFetch(key, field interface{}) ([]byte, error) {
	c := p.pool.Get()
	defer c.Close()

	b, err := redis.Bytes(c.Do("HGET", key, field))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	return b, nil
}

func CreateCache(server, auth string) c.Cache {
	REDIS_POOL = createPool(server, auth)
	return &cache{pool: REDIS_POOL}
}
