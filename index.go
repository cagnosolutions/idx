package idx

import (
	"bytes"
	"fmt"
	"strings"
)

const ORDER = 4

// node represents a tree's node
type node struct {
	numKeys int
	keys    [ORDER - 1][]byte
	ptrs    [ORDER]interface{}
	parent  *node
	isLeaf  bool
	next    *node
}

func (n *node) hasKey(key []byte) bool {
	for i := 0; i < n.numKeys; i++ {
		if bytes.Equal(key, n.keys[i]) {
			return true
		}
	}
	return false
}

// leaf node record
type Record struct {
	Key []byte
	Val int
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
func (t *Tree) Has(key []byte) bool {
	return t.Get(key) != nil
}

// Add inserts a new record using provided key.
// It only inserts if the key does not already exist.
func (t *Tree) Add(key []byte, val int) {
	// if the tree is empty
	if t.root == nil {
		t.root = startNewTree(key, &Record{key, val})
		return
	}
	// tree already exists, lets see what we
	// get when we try to find the correct leaf
	leaf := findLeaf(t.root, key)
	// ensure the leaf does not contain the key
	if leaf.hasKey(key) {
		return
	}
	// create record ptr for given value
	ptr := &Record{key, val}
	// tree already exists, and ready to insert into
	if leaf.numKeys < ORDER-1 {
		insertIntoLeaf(leaf, ptr.Key, ptr)
		return
	}
	// otherwise, insert, split, and balance... returning updated root
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, ptr.Key, ptr)
}

// Set is mainly used for re-indexing
// as it assumes the data to already
// be contained the tree/index. it will
// overwrite duplicate keys, as it does
// not check to see if the key exists...
func (t *Tree) Set(key []byte, val int) {
	// if the tree is empty, start a new one
	if t.root == nil {
		t.root = startNewTree(key, &Record{key, val})
		return
	}
	// create record ptr for given value
	ptr := &Record{key, val}
	// tree already exists, and ready to insert a non
	// duplicate value. find proper leaf to insert into
	leaf := findLeaf(t.root, ptr.Key)
	// if the leaf has room, then insert key and record
	if leaf.numKeys < ORDER-1 {
		insertIntoLeaf(leaf, ptr.Key, ptr)
		return
	}
	// otherwise, insert, split, and balance... returning updated root
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, ptr.Key, ptr)
}

/*
 *	inserting internals
 */

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
	if left.parent == nil {
		return insertIntoNewRoot(left, key, right)
	}
	leftIndex := getLeftIndex(left.parent, left)
	if left.parent.numKeys < ORDER-1 {
		return insertIntoNode(root, left.parent, leftIndex, key, right)
	}
	return insertIntoNodeAfterSplitting(root, left.parent, leftIndex, key, right)
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
func insertIntoNode(root, n *node, leftIndex int, key []byte, right *node) *node {
	/*for i := n.numKeys; i > leftIndex; i-- {
		n.ptrs[i+1] = n.ptrs[i]
		n.keys[i] = n.keys[i-1]
	}*/

	copy(n.ptrs[leftIndex+2:], n.ptrs[leftIndex+1:])
	copy(n.keys[leftIndex+1:], n.keys[leftIndex:])

	n.ptrs[leftIndex+1] = right
	n.keys[leftIndex] = key
	n.numKeys++
	return root
}

//var PtempKeys = sync.Pool{New: func() interface{} { return [ORDER][]byte{} }}
//var PtempPtrs = sync.Pool{New: func() interface{} { return [ORDER + 1]interface{}{} }}

// insert a new key, ptr to a node causing node to split
func insertIntoNodeAfterSplitting(root, oldNode *node, leftIndex int, key []byte, right *node) *node {
	var i, j int
	//var child *node
	//var prime []byte

	//tmpKeys := PtempKeys.Get().([ORDER][]byte)
	//tmpPtrs := PtempPtrs.Get().([ORDER + 1]interface{})

	var tmpKeys [ORDER][]byte
	var tmpPtrs [ORDER + 1]interface{}

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

	/*
		oldNode.numKeys = split
		copy(oldNode.keys[:split], tmpKeys[:split])
		copy(oldNode.ptrs[:split+1], tmpPtrs[:split+1])
	*/

	prime := tmpKeys[split-1]

	/*
		end := ORDER - (split + 1)
		copy(oldNode.keys[:end], tmpKeys[split+1:])
		copy(oldNode.ptrs[:end+1], tmpPtrs[split+2:])
		newNode.numKeys = end
	*/

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

	//PtempKeys.Put(tmpKeys)
	//PtempPtrs.Put(tmpPtrs)

	newNode.parent = oldNode.parent

	for i = 0; i <= newNode.numKeys; i++ {
		/*child = newNode.ptrs[i].(*node)
		child.parent = newNode*/

		newNode.ptrs[i].(*node).parent = newNode
	}
	return insertIntoParent(root, oldNode, prime, newNode)
}

/*
 *	Leaf node insert internals
 */

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

// Get returns the record for
// a given key if it exists
func (t *Tree) Get(key []byte) *Record {
	n := findLeaf(t.root, key)
	if n == nil {
		return nil
	}
	var i int
	for i = 0; i < n.numKeys; i++ {
		if bytes.Compare(n.keys[i], key) == 0 {
			break
		}
	}
	if i == n.numKeys {
		return nil
	}
	return n.ptrs[i].(*Record)
}

/*
 *	Get node internals
 */

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

// binary search utility
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

// breadth-first-search algorithm, kind of
func (t *Tree) BFS() {
	if t.root == nil {
		return
	}
	c, h := t.root, 0
	for !c.isLeaf {
		c = c.ptrs[0].(*node)
		h++
	}
	fmt.Printf(`[`)
	for h >= 0 {
		for i := 0; i < ORDER; i++ {
			if i == ORDER-1 && c.ptrs[ORDER-1] != nil {
				fmt.Printf(` -> `)
				c = c.ptrs[ORDER-1].(*node)
				i = 0
				continue
			}
			fmt.Printf(`[%s]`, c.keys[i])
		}
		fmt.Println()
		h--
	}
	fmt.Printf(`]\n`)
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

// Del deletes a record by key
func (t *Tree) Del(key []byte) {
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
func (t *Tree) All() []int {
	leaf := findFirstLeaf(t.root)
	if leaf == nil {
		return nil
	}
	var vals []int
	for {
		for i := 0; i < leaf.numKeys; i++ {
			if leaf.ptrs[i] != nil {
				// get record from leaf
				rec := leaf.ptrs[i].(*Record)
				// get doc and append to docs
				vals = append(vals, rec.Val)
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

/*
 * Printing methods
 */

var queue *node = nil

func enQueue(n *node) {
	var c *node
	if queue == nil {
		queue = n
		queue.next = nil
	} else {
		c = queue
		for c.next != nil {
			c = c.next
		}
		c.next = n
		n.next = nil
	}
}

func deQueue() *node {
	var n *node = queue
	queue = queue.next
	n.next = nil
	return n
}

func pathToRoot(root, child *node) int {
	var length int
	var c *node = child
	for c != root {
		c = c.parent
		length++
	}
	return length
}

func (t *Tree) String() string {
	var i, rank, newRank int
	if t.root == nil {
		return "[]"
	}
	queue = nil
	var tree string
	enQueue(t.root)
	tree = "[\n["
	for queue != nil {
		n := deQueue()
		if n.parent != nil && n == n.parent.ptrs[0] {
			newRank = pathToRoot(t.root, n)
			if newRank != rank {
				rank = newRank
				f := strings.LastIndex(tree, ",")
				tree = tree[:f] + tree[f+1:]
				tree += "],\n["
			}
		}
		tree += "["
		var keys []string
		for i = 0; i < n.numKeys; i++ {
			keys = append(keys, fmt.Sprintf("%q", n.keys[i]))
			//tree += fmt.Sprintf("%s", n.keys[i])
		}
		tree += strings.Join(keys, ",")
		if !n.isLeaf {
			for i = 0; i <= n.numKeys; i++ {
				enQueue(n.ptrs[i].(*node))
			}
		}
		tree += "],"
	}
	//tree[f] = "]"
	f := strings.LastIndex(tree, ",")
	tree = tree[:f] + tree[f+1:]
	tree += "]\n]"
	tree += "\n"
	return tree
}
