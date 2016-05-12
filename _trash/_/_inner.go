package adb

// insert a new key, ptr to a node
func insertIntoNode(root, n *node, leftIndex int, key []byte, right *node) *node {
	var i int
	for i = n.numKeys; i > leftIndex; i-- {
		n.ptrs[i+1] = n.ptrs[i]
		n.keys[i] = n.keys[i-1]
	}
	n.ptrs[leftIndex+1] = right
	n.keys[leftIndex] = key
	n.numKeys++
	return root
}

// insert a new key, ptr to a node causing node to split
func insertIntoNodeAfterSplitting(root, oldNode *node, leftIndex int, key []byte, right *node) *node {
	var i, j int
	var child *node
	var tmpKeys [ORDER][]byte
	var tmpPtrs [ORDER + 1]interface{}
	var prime []byte

	for i < oldNode.numKeys+1 {
		if j == leftIndex+1 {
			j++
		}
		tmpPtrs[j] = oldNode.ptrs[i]
		i++
		j++
	}

	i = 0
	j = 0

	for i < oldNode.numKeys {
		if j == leftIndex {
			j++
		}
		tmpKeys[j] = oldNode.keys[i]
		i++
		j++
	}

	tmpPtrs[leftIndex+1] = right
	tmpKeys[leftIndex] = key

	split := cut(ORDER)
	newNode := &node{}
	oldNode.numKeys = 0

	for i = 0; i < split-1; i++ {
		oldNode.ptrs[i] = tmpPtrs[i]
		oldNode.keys[i] = tmpKeys[i]
		oldNode.numKeys++
	}

	oldNode.ptrs[i] = tmpPtrs[i]
	prime = tmpKeys[split-1]

	j = 0
	for i++; i < ORDER; i++ {
		newNode.ptrs[j] = tmpPtrs[i]
		newNode.keys[j] = tmpKeys[i]
		newNode.numKeys++
		j++
	}

	newNode.ptrs[j] = tmpPtrs[i]

	// free tmps...
	for i = 0; i < ORDER; i++ {
		tmpKeys[i] = nil
		tmpPtrs[i] = nil
	}
	tmpPtrs[ORDER] = nil

	newNode.parent = oldNode.parent

	for i = 0; i <= newNode.numKeys; i++ {
		child = newNode.ptrs[i].(*node)
		child.parent = newNode
	}
	return insertIntoParent(root, oldNode, prime, newNode)
}
