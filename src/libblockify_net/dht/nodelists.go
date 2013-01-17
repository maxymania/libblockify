package dht

import "container/heap"
import "net"
import "sync"

type Node struct{
	Id       []byte
	Addr     net.Addr
}

type nlEntry struct{
	Distance int
	Id       []byte
	Addr     net.Addr
}

type nlHeap struct{
	A []*nlEntry
}
func (h nlHeap) Push(x interface{}) {
	h.A = append(h.A,x.(*nlEntry))
}
func (h nlHeap) Pop() (r interface{}) {
	l := len(h.A)-1
	if l<0 { return }
	r = h.A[l]
	h.A=h.A[:l]
	return
}
func (h nlHeap) Len() int {
	return len(h.A)
}
func (h nlHeap) Less(i, j int) bool {
	return h.A[i].Distance < h.A[j].Distance
}
func (h nlHeap) Swap(i, j int) {
	t := h.A[i]
	h.A[i] = h.A[j]
	h.A[j] = t
}


type Nlist struct{
	OwnNodeID []byte
	MaxLen int
	heap *nlHeap
	index map[string]bool
	mutex sync.RWMutex
}
func (n *Nlist) Constructor() *Nlist {
	n.heap = &nlHeap{[]*nlEntry{}}
	return n
}
func (n *Nlist) Len() int { return n.heap.Len() }
func (n *Nlist) InsertNode(node *Node) bool {
	return n.Insert(node.Id,node.Addr)
}
func (n *Nlist) Insert(id []byte, addr net.Addr) bool {
	n.mutex.Lock() ; defer n.mutex.Unlock()
	_,exists := n.index[string(id)]
	if exists { return false }
	d := HammingDistance(id,n.OwnNodeID)
	heap.Push(n.heap,&nlEntry{d,id,addr})
	if n.heap.Len()>n.MaxLen { n.iremove() }
	return true
}
func (n *Nlist) Remove() *Node {
	nd := n.iremove()
	return &Node{nd.Id,nd.Addr}
}
func (n *Nlist) iremove() *nlEntry {
	n.mutex.Lock() ; defer n.mutex.Unlock()
	nd := heap.Pop(n.heap).(*nlEntry)
	delete(n.index,string(nd.Id))
	return nd
}
func (n *Nlist) Iterate(dst chan *Node) {
	n.mutex.RLock() ; defer n.mutex.RUnlock()
	for _,nd := range n.heap.A {
		dst <- &Node{nd.Id,nd.Addr}
	}
}


type NodeMap map[string]*Node
func (nm NodeMap) InsertNode(n *Node) bool{
	id := string(n.Id)
	_,ok := nm[id]
	if ok { return false }
	nm[id]=n
	return true
}
func (nm NodeMap) RemoveNode(n *Node){
	id := string(n.Id)
	_,ok := nm[id]
	if ok {	delete(nm,id) }
}
func (nm NodeMap) Iterate(dst chan *Node) {
	for _,n := range nm {
		dst <- n
	}
}

