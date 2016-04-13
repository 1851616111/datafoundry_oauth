package main

import "errors"

var ErrNotFound = errors.New("not found")

func NotFount(err error) bool {
	if err == ErrNotFound {
		return true
	}
	return false
}
