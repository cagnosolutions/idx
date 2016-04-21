package idx

// ORDER is defined as the maximum number of pointers in any given node
// MIN_ORDER <= ORDER <= MAX_ORDER
// internal node min ptrs = ORDER/2 round up
// internal node max ptrs = ORDER
// leaf node min ptrs (ORDER-1)/ round up
// leaf node max ptrs ORDER-1

// node represents a tree's node
type node struct {
	numKeys int
	keys    [ORDER - 1]Key
	ptrs    [ORDER]interface{}
	parent  *node
	isLeaf  bool
}

// leaf node record
type record struct {
	key Key
	val Val
}

// Tree represents the main b+tree
type Tree struct {
	root *node
}

// NewTree creates and returns a new tree
func NewTree() *Tree {
	return &Tree{}
}

// Has returns a boolean indicating weather or not the
// provided key and associated record / value exists.
func (t *Tree) Has(key Key) bool {
	return getRecord(t.root, key) != nil
}

// Add inserts a new record / value using provided key.
// It only inserts if the key does not already exist.
func (t *Tree) Add(key Key, value Val) {
	// ignore duplicates: if a value
	// can be found for a given key,
	// simply return, don't insert
	if t.Get(key) != nil {
		return
	}
	// otherwise simply call set
	t.Set(key, value)
}

// Put is mainly used for re-indexing
// as it assumes the data to already
// be contained on the disk, so it's
// policy is to just forcefully put it
// in the btree index. it will overwrite
// duplicate keys, as it does not check
// to see if the key already exists...
func (t *Tree) Put(key Key, value Val) {
	// create record ptr for given value
	ptr := &record{key, value}
	// if the tree is empty, start a new one
	if t.root == nil {
		t.root = startNewTree(ptr.key, ptr)
		return
	}
	// tree already exists, and ready to insert a non
	// duplicate value. find proper leaf to insert into
	leaf := findLeaf(t.root, ptr.key)
	// if the leaf has room, then insert key and record
	if leaf.numKeys < ORDER-1 {
		insertIntoLeaf(leaf, ptr.key, ptr)
		return
	}
	// otherwise, insert, split, and balance... returning updated root
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, ptr.key, ptr)
}

// Set volatility inserts or updates a value based on provided key
func (t *Tree) Set(key Key, value Val) {
	// don't ignore duplicates: if
	// a value can be found for a
	// given key, simply update the
	// record value and return
	if r := getRecord(t.root, key); r != nil {
		// update
		r.val = value
		return
	}
	// create record ptr for given value
	ptr := &record{key, value}
	// if the tree is empty, start a new one
	if t.root == nil {
		t.root = startNewTree(ptr.key, ptr)
		return
	}
	// tree already exists, and ready to insert a non
	// duplicate value. find proper leaf to insert into
	leaf := findLeaf(t.root, ptr.key)
	// if the leaf has room, then insert key and record
	if leaf.numKeys < ORDER-1 {
		insertIntoLeaf(leaf, ptr.key, ptr)
		return
	}
	// otherwise, insert, split, and balance... returning updated root
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, ptr.key, ptr)
}

/*
 *	Add, Put, Set inserting internals
 */

// first insertion, start a new tree
func startNewTree(key Key, ptr *record) *node {
	root := &node{isLeaf: true}
	root.keys[0] = key
	root.ptrs[0] = ptr
	root.ptrs[ORDER-1] = nil
	root.parent = nil
	root.numKeys++
	return root
}

// creates a new root for two sub-trees and inserts the key into the new root
func insertIntoNewRoot(left *node, key Key, right *node) *node {
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
func insertIntoParent(root, left *node, key Key, right *node) *node {
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
func getLeftIndex(parent, left *node) int {
	var leftIndex int
	for leftIndex <= parent.numKeys && parent.ptrs[leftIndex] != left {
		leftIndex++
	}
	return leftIndex
}

/*
 *	Inner node insert internals
 */

// insert a new key, ptr to a node
func insertIntoNode(root, n *node, leftIndex int, key Key, right *node) *node {
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
func insertIntoNodeAfterSplitting(root, oldNode *node, leftIndex int, key Key, right *node) *node {
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

/*
 *	Leaf node insert internals
 */

// inserts a new key and *record into a leaf, then returns leaf
func insertIntoLeaf(leaf *node, key Key, ptr *record) {
	var i, insertionPoint int
	for insertionPoint < leaf.numKeys && Compare(leaf.keys[insertionPoint], key) == -1 {
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
func insertIntoLeafAfterSplitting(root, leaf *node, key Key, ptr *record) *node {
	// perform linear search to find index to insert new record
	var insertionIndex int
	for insertionIndex < ORDER-1 && Compare(leaf.keys[insertionIndex], key) == -1 {
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

// Get finds a record for a given key
func (t *Tree) Get(key Key) Val {
	r := getRecord(t.root, key)
	if r != nil {
		return r.val
	}
	return *new(Val)
}

/*
 *	Get node internals
 */

// find leaf type node for a given key
func findLeaf(n *node, key Key) *node {
	if n == nil {
		return n
	}
	for !n.isLeaf {
		n = n.ptrs[search(n, key)].(*node)
	}
	return n
}

// binary search utility
func search(n *node, key Key) int {
	lo, hi := 0, n.numKeys-1
	for lo <= hi {
		md := (lo + hi) >> 1
		switch cmp := Compare(key, n.keys[md]); {
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

// finds the first leaf in the tree (lexicographically)
func findFirstLeaf(root *node) *node {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = c.ptrs[0].(*node)
	}
	return c
}

// getRecord returns the record for
// a given key if it exists
func getRecord(root *node, key Key) *record {
	n := findLeaf(root, key)
	if n == nil {
		return nil
	}
	var i int
	for i = 0; i < n.numKeys; i++ {
		if Equal(n.keys[i], key) {
			break
		}
	}
	if i == n.numKeys {
		return nil
	}
	return n.ptrs[i].(*record)
}

// Del deletes a record by key
func (t *Tree) Del(key Key) {
	record := t.Get(key)
	leaf := findLeaf(t.root, key)
	if record != nil && leaf != nil {
		// remove from tree, and rebalance
		t.root = deleteEntry(t.root, leaf, key, record)
	}
}

/*
 * Delete internals
 */

// helper for delete methods... returns index of
// a nodes nearest sibling to the left if one exists
func getNeighborIndex(n *node) int {
	for i := 0; i <= n.parent.numKeys; i++ {
		if n.parent.ptrs[i] == n {
			return i - 1
		}
	}
	panic("Search for nonexistent ptr to node in parent.")
	return 1
}

func removeEntryFromNode(n *node, key Key, ptr interface{}) *node {
	var i, numPtrs int
	// remove key and shift over keys accordingly
	for Compare(n.keys[i], key) != 0 {
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

// All returns all of the values in the tree (lexicographically)
func (t *Tree) All() []Val {
	leaf := findFirstLeaf(t.root)
	if leaf == nil {
		return nil
	}
	var vals []Val
	for {
		for i := 0; i < leaf.numKeys; i++ {
			if leaf.ptrs[i] != nil {
				// get record from leaf
				rec := leaf.ptrs[i].(*record)
				// get doc and append to docs
				vals = append(vals, rec.val)
			}
		}
		// we're at the end, no more leaves to iterate
		if leaf.ptrs[ORDER-1] == nil {
			break
		}
		// increment/follow pointer to next leaf node
		leaf = leaf.ptrs[ORDER-1].(*node)
	}
	return vals
}

// Count returns the number of records in the tree
func (t *Tree) Count() int {
	if t.root == nil {
		return -1
	}
	c := t.root
	for !c.isLeaf {
		c = c.ptrs[0].(*node)
	}
	var size int
	for {
		size += c.numKeys
		if c.ptrs[ORDER-1] != nil {
			c = c.ptrs[ORDER-1].(*node)
		} else {
			break
		}
	}
	return size
}

// Close destroys all the nodes of the tree
func (t *Tree) Close() {
	destroyTreeNodes(t.root)
}

// cut will return the proper
// split point for a node
func cut(length int) int {
	if length%2 == 0 {
		return length / 2
	}
	return length/2 + 1
}