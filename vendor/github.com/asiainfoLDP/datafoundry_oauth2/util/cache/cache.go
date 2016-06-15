package cache

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Cache interface {
	HCache(key, field interface{}, value []byte) error
	HFetch(key, field interface{}) ([]byte, error)
}

type CacheMan interface {
	HCacheObject(key, field interface{}, Object interface{}) error
	HFetchObject(key, field interface{}, Object interface{}) error
}

type cache struct {
	Cache
}

func (c *cache) HCacheObject(key, field interface{}, obj interface{}) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.HCache(key, field, b)
}

func (c *cache) HFetchObject(key, field interface{}, obj interface{}) error {
	if err := validateObject(obj); err != nil {
		return err
	}

	b, err := c.HFetch(key, field)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, obj)
}

func NewCacheMan(c Cache) CacheMan {
	return &cache{c}
}

func validateObject(obj interface{}) error {
	k := reflect.TypeOf(obj).Kind()
	switch k {
	case reflect.Ptr, reflect.Map:
		return nil
	default:
		return fmt.Errorf("not allow object type %v", k)
	}
}
