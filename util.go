package main

import "os"

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if value == "" && required {
		panic("no exist env " + key)
	}

	return value
}
