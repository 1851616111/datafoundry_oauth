package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"reflect"
)

const (
	EtcdUserRegistry = "/oauth/users"
)

type storeConfig struct {
	Addr, Port, User, Passwd string
}

func (c *storeConfig) newClient() Store {
	cli := etcd.NewClient([]string{fmt.Sprintf("%s:%s", c.Addr, c.Port)})
	cli.SetCredentials(c.User, c.Passwd)
	return &Etcd{cli}
}

type Store interface {
	set(key string, value interface{}) error
	get(key string, sort, recursive bool) (string, error)
}

type Etcd struct {
	*etcd.Client
}

func (c *Etcd) set(key string, value interface{}) error {
	t := reflect.TypeOf(value).Kind()
	switch t {
	case reflect.String:
		c.Set(key, value.(string), 0)
	case reflect.Struct, reflect.Ptr, reflect.Map:
		if b, err := json.Marshal(value); err != nil {
			return err
		} else {
			c.Set(key, string(b), 0)
		}
	default:
		return errors.New(fmt.Sprintf("unsupport value type %s", t.String()))
	}

	return nil
}

func (c *Etcd) get(key string, sort, recursive bool) (string, error) {
	rsp, err := c.Get(key, sort, recursive)
	if err != nil {
		return "", err
	}

	return rsp.Node.Value, nil
}

func getJson(key string, box interface{}) error {
	b, err := db.get(key, true, false)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(b), box)
}
