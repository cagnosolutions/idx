package adb

import "bytes"

// find leaf type node for a given key
func findLeaf(n *node, key []byte) *node {
	if n == nil {
		return n
	}
	for !n.isLeaf {
		n = n.ptrs[search(n, key)].(*node)
	}
	return n
}

func search(n *node, key []byte) int {
	lo, hi := 0, n.numKeys-1
	for lo <= hi {
		md := (lo + hi) >> 1
		switch cmp := bytes.Compare(key, n.keys[md]); {
		case cmp > 0:
			lo = md + 1
		case cmp == 0:
			return md
		default:
			hi = md - 1
		}
	}
	return lo
}
