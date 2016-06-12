package etcd

import (
	"github.com/coreos/go-etcd/etcd"
)

func RangeNodeFunc(node *etcd.Node, fn func(*etcd.Node)) {
	if !isDir(node) {
		fn(node)
		return
	}

	for _, child := range node.Nodes {
		if isDir(child) {
			go RangeNodeFunc(child, fn)
		} else {
			fn(child)
		}
	}
}

func isDir(n *etcd.Node) bool {
	return n.Dir
}
