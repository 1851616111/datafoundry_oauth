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
	getValue(key string) (string, error)
	getDir(key string) (*etcd.Response, error)
	delete(key string, recursive bool) error
}

type Etcd struct {
	*etcd.Client
}

func (c *Etcd) set(key string, value interface{}) error {
	t := reflect.TypeOf(value).Kind()

	switch t {
	case reflect.String:
		return notReachErrRetry(func(c *Etcd) error {
			_, err := c.Set(key, value.(string), 0)
			return err
		})
	case reflect.Struct, reflect.Ptr, reflect.Map:
		if b, err := json.Marshal(value); err != nil {
			return err
		} else {
			return notReachErrRetry(func(c *Etcd) error {
				_, err := c.Set(key, string(b), 0)
				return err
			})
		}
	default:
		return errors.New(fmt.Sprintf("unsupport value type %s", t.String()))
	}

	return nil
}

func (c *Etcd) getValue(key string) (string, error) {
	var rsp *etcd.Response
	var err error

	err = notReachErrRetry(func(c *Etcd) error {
		rsp, err = c.Get(key, true, false)
		return err
	})

	if err != nil {
		return "", err
	}

	return rsp.Node.Value, nil
}

func (c *Etcd) getDir(key string) (*etcd.Response, error) {
	var rsp *etcd.Response
	var err error

	err = notReachErrRetry(func(c *Etcd) error {
		rsp, err = c.Get(key, true, true)
		return err
	})

	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (c *Etcd) delete(key string, recursive bool) error {
	var rsp *etcd.Response
	var err error

	err = notReachErrRetry(func(c *Etcd) error {
		rsp, err = c.Delete(key, recursive)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}

func notReachErrRetry(f func(c *Etcd) error) (err error) {
	err = f(db.(*Etcd))

	if isEtcdNotReachableErr(err) {
		refreshDB()
		err = f(db.(*Etcd))

		if isEtcdNotReachableErr(err) {
			err = errors.New("Server Internal Error")
			return
		}
	}

	return
}

func isEtcdNotReachableErr(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*etcd.EtcdError); ok && e.ErrorCode == etcd.ErrCodeEtcdNotReachable {
		return true
	}

	return false
}

func refreshDB() {
	db = dbConf.newClient()
}
