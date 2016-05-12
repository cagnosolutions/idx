package adb

// first insertion, start a new tree
func startNewTree(key []byte, ptr *Record) *node {
	root := &node{isLeaf: true}
	root.keys[0] = key
	root.ptrs[0] = ptr
	root.ptrs[ORDER-1] = nil
	root.parent = nil
	root.numKeys++
	return root
}

// creates a new root for two sub-trees and inserts the key into the new root
func insertIntoNewRoot(left *node, key []byte, right *node) *node {
	root := &node{}
	root.keys[0] = key
	root.ptrs[0] = left
	root.ptrs[1] = right
	root.numKeys++
	root.parent = nil
	left.parent = root
	right.parent = root
	return root
}

// insert a new node (leaf or internal) into tree, return root of tree
func insertIntoParent(root, left *node, key []byte, right *node) *node {
	var leftIndex int
	var parent *node
	parent = left.parent
	if parent == nil {
		return insertIntoNewRoot(left, key, right)
	}
	leftIndex = getLeftIndex(parent, left)
	if parent.numKeys < ORDER-1 {
		return insertIntoNode(root, parent, leftIndex, key, right)
	}
	return insertIntoNodeAfterSplitting(root, parent, leftIndex, key, right)
}

// helper->insert_into_parent
// used to find index of the parent's ptr to the
// node to the left of the key to be inserted
// NOTE: best
func getLeftIndex(parent, left *node) int {
	var leftIndex int
	for leftIndex <= parent.numKeys && parent.ptrs[leftIndex] != left {
		leftIndex++
	}
	return leftIndex
}
