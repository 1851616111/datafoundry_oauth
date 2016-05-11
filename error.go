package main

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

func NotFound(err error) bool {
	if err == ErrNotFound {
		return true
	}
	return false
}

var etcdErrKeyNotFound = errors.New("Key not found")

func EtcdKeyNotFound(err error) bool {
	if strings.Contains(err.Error(), etcdErrKeyNotFound.Error()) {
		return true
	}
	return false
}
