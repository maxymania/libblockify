package relaxdht

import (
	"net"
	"sort"
	"sync"
	"fmt"
	"encoding/hex"
)

type Contact struct{
	Distance int
	Id string
	Addr net.Addr
}

type ContactTable struct{
	mutex sync.Mutex
	Data []Contact
}
func (c *ContactTable) Len() int { return len(c.Data) }
func (c *ContactTable) Less(i, j int) bool { return c.Data[i].Id<c.Data[j].Id }
func (c *ContactTable) Swap(i, j int) {
	temp := c.Data[i]
	c.Data[i] = c.Data[j]
	c.Data[j] = temp
}
func (c *ContactTable) Sort() {
	sort.Sort(c)
}
func (c *ContactTable) SearchId(id string) (bool,int){
	for i,cc := range c.Data {
		if cc.Id==id { return true,i }
	}
	return false,0
}
func (c *ContactTable) GetCloser(ids string,od int) <- chan net.Addr{
	lower := make(chan net.Addr)
	id,e := hex.DecodeString(ids)
	if e!=nil||len(id)!=64 { close(lower) ; return lower }
	go func(){
		for _,cc := range c.Data {
			dist,e := HammingDistanceRH(id,cc.Id)
			if e!=nil { continue }
			if od>dist {
				lower <- cc.Addr
			}
		}
		close(lower)
	}()
	return lower
}
func (c *ContactTable) GetClosest(id []byte) (Contact,int,bool){
	var cont Contact
	ok := false
	curdist := 0
	for _,cc := range c.Data {
		dist,e := HammingDistanceRH(id,cc.Id)
		if e!=nil { continue }
		if ok || curdist>dist {
			cc=cont
			curdist=dist
			ok=true
		}
	}
	return cont,curdist,ok
}
func (c *ContactTable) Insert(con Contact){
	c.mutex.Lock() ; defer c.mutex.Unlock()
	c.Data = append(c.Data,con)
}
func (c *ContactTable) Debug() {
	// c.mutex.Lock() ; defer c.mutex.Unlock()
	for _,cc := range c.Data {
		fmt.Println(cc.Id,cc.Addr,"->",cc.Distance)
	}
}
func (c *ContactTable) Limit(n int){
	if len(c.Data)>n { c.Data = c.Data[:n] }
}
func (c *ContactTable) SLimit(n int){
	if len(c.Data)>n { c.Sort() ; c.Data = c.Data[:n] }
}

