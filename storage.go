package main

type store interface {
	set(key, value string) error
	get(key string) (string, error)
}
