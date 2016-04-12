package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if value == "" && required {
		panic("no exist env " + key)
	}

	return value
}

func httpAddrMaker(addr string) string {
	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if !strings.HasPrefix(addr, "http://") {
		return fmt.Sprintf("http://%s", addr)
	}

	return addr
}

func headers(r *http.Request, keys ...string) map[string]string {
	m := map[string]string{}
	for i := 0; i < len(keys); i++ {
		if value := r.Header.Get(keys[i]); value != "" {
			m[keys[i]] = value
		}
	}

	return m
}

func printConfig(c *storeConfig) {
	fmt.Printf("[ETCD_HTTP_ADDR]=%s\n", c.Addr)
	fmt.Printf("[ETCD_HTTP_PORT]=%s\n", c.Port)
	fmt.Printf("[ETCD_USER]=%s\n", c.User)
	fmt.Printf("[ETCD_PASSWOR]=%s\n", c.Passwd)
}
