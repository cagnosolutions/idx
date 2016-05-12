package adb

import "bytes"

// inserts a new key and *record into a leaf, then returns leaf
func insertIntoLeaf(leaf *node, key []byte, ptr *Record) {
	var i, insertionPoint int
	for insertionPoint < leaf.numKeys && bytes.Compare(leaf.keys[insertionPoint], key) == -1 {
		insertionPoint++
	}
	for i = leaf.numKeys; i > insertionPoint; i-- {
		leaf.keys[i] = leaf.keys[i-1]
		leaf.ptrs[i] = leaf.ptrs[i-1]
	}
	leaf.keys[insertionPoint] = key
	leaf.ptrs[insertionPoint] = ptr
	leaf.numKeys++
}

// inserts a new key and *record into a leaf, so as
// to exceed the order, causing the leaf to be split
func insertIntoLeafAfterSplitting(root, leaf *node, key []byte, ptr *Record) *node {
	// perform linear search to find index to insert new record
	var insertionIndex int
	for insertionIndex < ORDER-1 && bytes.Compare(leaf.keys[insertionIndex], key) == -1 {
		insertionIndex++
	}
	var tmpKeys [ORDER][]byte
	var tmpPtrs [ORDER]interface{}
	var i, j int
	// copy leaf keys & ptrs to temp
	// reserve space at insertion index for new record
	for i < leaf.numKeys {
		if j == insertionIndex {
			j++
		}
		tmpKeys[j] = leaf.keys[i]
		tmpPtrs[j] = leaf.ptrs[i]
		i++
		j++
	}
	tmpKeys[insertionIndex] = key
	tmpPtrs[insertionIndex] = ptr

	leaf.numKeys = 0
	// index where to split leaf
	split := cut(ORDER - 1)
	// over write original leaf up to split point
	for i = 0; i < split; i++ {
		leaf.ptrs[i] = tmpPtrs[i]
		leaf.keys[i] = tmpKeys[i]
		leaf.numKeys++
	}
	// create new leaf
	newLeaf := &node{isLeaf: true}
	// writing to new leaf from split point to end of giginal leaf pre split
	j = 0
	for i = split; i < ORDER; i++ {
		newLeaf.ptrs[j] = tmpPtrs[i]
		newLeaf.keys[j] = tmpKeys[i]
		newLeaf.numKeys++
		j++
	}
	// freeing tmps...
	for i = 0; i < ORDER; i++ {
		tmpPtrs[i] = nil
		tmpKeys[i] = nil
	}
	newLeaf.ptrs[ORDER-1] = leaf.ptrs[ORDER-1]
	leaf.ptrs[ORDER-1] = newLeaf
	for i = leaf.numKeys; i < ORDER-1; i++ {
		leaf.ptrs[i] = nil
	}
	for i = newLeaf.numKeys; i < ORDER-1; i++ {
		newLeaf.ptrs[i] = nil
	}
	newLeaf.parent = leaf.parent
	return insertIntoParent(root, leaf, newLeaf.keys[0], newLeaf)
}
