package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"reflect"
)

const (
	Etcd_Sub_Root_Registry = "/oauth/namespace"
)

type storeConfig struct {
	Addr, Port, User, Passwd string
}

func (c *storeConfig) newClient() client {
	return client{etcd.NewClient([]string{fmt.Sprintf("%s:%s", c.Addr, c.Port)})}
}

type Store interface {
	set(key, value string) error
	namespaceSet(namespace, key string, value interface{}) error
	get(key string) (string, error)
}

type client struct {
	*etcd.Client
}

func (c *client) set(key string, value interface{}) error {
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

func (c *client) get(key string, sort, recursive bool) (string, error) {
	rsp, err := c.Get(key, sort, recursive)
	if err != nil {
		return "", err
	}

	return rsp.Node.Value, nil
}

func (c *client) namespaceSet(namespace, key string, value interface{}) error {
	path := fmt.Sprintf("%s/%s/%s", Etcd_Sub_Root_Registry, namespace, key)
	return etcdClient.set(path, value)
}
