package adb

import (
	"bytes"
	"log"
)

// helper for delete methods... returns index of
// a nodes nearest sibling to the left if one exists
func getNeighborIndex(n *node) int {
	for i := 0; i <= n.parent.numKeys; i++ {
		if n.parent.ptrs[i] == n {
			return i - 1
		}
	}
	log.Fatalf("Search for nonexistent ptr to node in parent.\nNode: %p\n", n)
	return 1
}

func removeEntryFromNode(n *node, key []byte, ptr interface{}) *node {
	var i, numPtrs int
	// remove key and shift over keys accordingly
	for !bytes.Equal(n.keys[i], key) {
		i++
	}
	for i++; i < n.numKeys; i++ {
		n.keys[i-1] = n.keys[i]
	}
	// remove ptr and shift other ptrs accordingly
	// first determine the number of ptrs
	if n.isLeaf {
		numPtrs = n.numKeys
	} else {
		numPtrs = n.numKeys + 1
	}
	i = 0
	for n.ptrs[i] != ptr {
		i++
	}

	for i++; i < numPtrs; i++ {
		n.ptrs[i-1] = n.ptrs[i]
	}
	// one key has been removed
	n.numKeys--
	// set other ptrs to nil for tidiness; remember leaf
	// nodes use the last ptr to point to the next leaf
	if n.isLeaf {
		for i := n.numKeys; i < ORDER-1; i++ {
			n.ptrs[i] = nil
		}
	} else {
		for i := n.numKeys + 1; i < ORDER; i++ {
			n.ptrs[i] = nil
		}
	}
	return n
}

// deletes an entry from the tree; removes record, key, and ptr from leaf and rebalances tree
func deleteEntry(root, n *node, key []byte, ptr interface{}) *node {
	var primeIndex, capacity int
	var neighbor *node
	var prime []byte

	// remove key, ptr from node
	n = removeEntryFromNode(n, key, ptr)

	if n == root {
		return adjustRoot(root)
	}

	var minKeys int
	// case: delete from inner node
	if n.isLeaf {
		minKeys = cut(ORDER - 1)
	} else {
		minKeys = cut(ORDER) - 1
	}
	// case: node stays at or above min order
	if n.numKeys >= minKeys {
		return root
	}

	// case: node is bellow min order; coalescence or redistribute
	neighborIndex := getNeighborIndex(n)
	if neighborIndex == -1 {
		primeIndex = 0
	} else {
		primeIndex = neighborIndex
	}
	prime = n.parent.keys[primeIndex]
	if neighborIndex == -1 {
		neighbor = n.parent.ptrs[1].(*node)
	} else {
		neighbor = n.parent.ptrs[neighborIndex].(*node)
	}
	if n.isLeaf {
		capacity = ORDER
	} else {
		capacity = ORDER - 1
	}

	// coalescence
	if neighbor.numKeys+n.numKeys < capacity {
		return coalesceNodes(root, n, neighbor, neighborIndex, prime)
	}
	return redistributeNodes(root, n, neighbor, neighborIndex, primeIndex, prime)
}

func adjustRoot(root *node) *node {
	// if non-empty root key and ptr
	// have already been deleted, so
	// nothing to be done here
	if root.numKeys > 0 {
		return root
	}
	var newRoot *node
	// if root is empty and has a child
	// promote first (only) child as the
	// new root node. If it's a leaf then
	// the while tree is empty...
	if !root.isLeaf {
		newRoot = root.ptrs[0].(*node)
		newRoot.parent = nil
	} else {
		newRoot = nil
	}
	root = nil // free root
	return newRoot
}

// merge (underflow)
func coalesceNodes(root, n, neighbor *node, neighborIndex int, prime []byte) *node {
	var i, j, neighborInsertionIndex, nEnd int
	var tmp *node
	// swap neight with node if nod eis on the
	// extreme left and neighbor is to its right
	if neighborIndex == -1 {
		tmp = n
		n = neighbor
		neighbor = tmp
	}
	// starting index for merged pointers
	neighborInsertionIndex = neighbor.numKeys
	// case nonleaf node, append k_prime and the following ptr.
	// append all ptrs and keys for the neighbors.
	if !n.isLeaf {
		// append k_prime (key)
		neighbor.keys[neighborInsertionIndex] = prime
		neighbor.numKeys++
		nEnd = n.numKeys
		i = neighborInsertionIndex + 1
		j = 0
		for j < nEnd {
			neighbor.keys[i] = n.keys[j]
			neighbor.ptrs[i] = n.ptrs[j]
			neighbor.numKeys++
			n.numKeys--
			i++
			j++
		}
		neighbor.ptrs[i] = n.ptrs[j]
		for i = 0; i < neighbor.numKeys+1; i++ {
			tmp = neighbor.ptrs[i].(*node)
			tmp.parent = neighbor
		}
	} else {
		// in a leaf; append the keys and ptrs.
		i = neighborInsertionIndex
		j = 0
		for j < n.numKeys {
			neighbor.keys[i] = n.keys[j]
			neighbor.ptrs[i] = n.ptrs[j]
			neighbor.numKeys++
			i++
			j++
		}
		neighbor.ptrs[ORDER-1] = n.ptrs[ORDER-1]
	}
	root = deleteEntry(root, n.parent, prime, n)
	n = nil // free n
	return root
}

// merge / redistribute
func redistributeNodes(root, n, neighbor *node, neighborIndex, primeIndex int, prime []byte) *node {
	var i int
	var tmp *node

	// case: node n has a neighnor to the left
	if neighborIndex != -1 {
		if !n.isLeaf {
			n.ptrs[n.numKeys+1] = n.ptrs[n.numKeys]
		}
		for i = n.numKeys; i > 0; i-- {
			n.keys[i] = n.keys[i-1]
			n.ptrs[i] = n.ptrs[i-1]
		}
		if !n.isLeaf {
			n.ptrs[0] = neighbor.ptrs[neighbor.numKeys]
			tmp = n.ptrs[0].(*node)
			tmp.parent = n
			neighbor.ptrs[neighbor.numKeys] = nil
			n.keys[0] = prime
			n.parent.keys[primeIndex] = neighbor.keys[neighbor.numKeys-1]
		} else {
			n.ptrs[0] = neighbor.ptrs[neighbor.numKeys-1]
			neighbor.ptrs[neighbor.numKeys-1] = nil
			n.keys[0] = neighbor.keys[neighbor.numKeys-1]
			n.parent.keys[primeIndex] = n.keys[0]
		}
	} else {
		// case: n is left most child (n has no left neighbor)
		if n.isLeaf {
			n.keys[n.numKeys] = neighbor.keys[0]
			n.ptrs[n.numKeys] = neighbor.ptrs[0]
			n.parent.keys[primeIndex] = neighbor.keys[1]
		} else {
			n.keys[n.numKeys] = prime
			n.ptrs[n.numKeys+1] = neighbor.ptrs[0]
			tmp = n.ptrs[n.numKeys+1].(*node)
			tmp.parent = n
			n.parent.keys[primeIndex] = neighbor.keys[0]
		}
		for i = 0; i < neighbor.numKeys-1; i++ {
			neighbor.keys[i] = neighbor.keys[i+1]
			neighbor.ptrs[i] = neighbor.ptrs[i+1]
		}
		if !n.isLeaf {
			neighbor.ptrs[i] = neighbor.ptrs[i+1]
		}
	}

	n.numKeys++
	neighbor.numKeys--
	return root
}

/*func destroy_tree(n *node) {
	destroy_tree_nodes(n)
}*/

func destroyTreeNodes(n *node) {
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			n.ptrs[i] = nil
		}
	} else {
		for i := 0; i < n.numKeys+1; i++ {
			destroyTreeNodes(n.ptrs[i].(*node))
		}
	}
	n = nil // free
}
