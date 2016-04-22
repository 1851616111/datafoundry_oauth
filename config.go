package main

import "fmt"

func printConfig(c *storeConfig) {
	fmt.Printf("[ETCD_HTTP_ADDR]=%s\n", c.Addr)
	fmt.Printf("[ETCD_HTTP_PORT]=%s\n", c.Port)
	fmt.Printf("[ETCD_USER]=%s\n", c.User)
	fmt.Printf("[ETCD_PASSWOR]=%s\n", c.Passwd)
}