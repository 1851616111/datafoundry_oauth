package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
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

type Cache interface {
	HCache(key, field interface{}, value []byte) error
	HFetch(key, field interface{}) ([]byte, error)
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

func CreateCache(server, auth string) Cache {
	REDIS_POOL = createPool(server, auth)
	return &cache{pool: REDIS_POOL}
}

func GetRedisMasterAddr(sentinelAddr string) (string, string) {
	if len(sentinelAddr) == 0 {
		log.Printf("Redis sentinelAddr is nil.")
		return "", ""
	}

	conn, err := redis.DialTimeout("tcp", sentinelAddr, time.Second*10, time.Second*10, time.Second*10)
	if err != nil {
		log.Printf("redis dial timeout(\"tcp\", \"%s\", %d) error(%v)", sentinelAddr, time.Second, err)
		return "", ""
	}
	defer conn.Close()

	redisMasterPair, err := redis.Strings(conn.Do("SENTINEL", "get-master-addr-by-name", "mymaster"))
	if err != nil {
		log.Printf("conn.Do(\"SENTINEL\", \"get-master-addr-by-name\", \"%s\") error(%v)", "mymaster", err)
		return "", ""
	}

	log.Printf("get redis addr: \"%v\"", redisMasterPair)
	if len(redisMasterPair) != 2 {
		return "", ""
	}
	return redisMasterPair[0], redisMasterPair[1]
}
