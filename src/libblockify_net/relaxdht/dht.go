package relaxdht

import (
	"fmt"
	"net"
	// "sync"
	"encoding/json"
	"encoding/hex"
	"strings"
	"strconv"
	"libblockify/bucket"
)

type Packet struct{
	Id string `json:"id"`
	Q string `json:"q"`
	OtherId string `json:"other-id,omitempty"`
	Appendix interface{} `json:"appendix,omitempty"`
}

type IHaveMsg struct{
	Id []byte
	Addr *net.TCPAddr
}

func appendixToUdpAddress(d interface{}) (net.Addr,error){
	s,ok := d.(string)
	if !ok { return nil,nil }
	addr,e := net.ResolveUDPAddr("udp",s)
	if e!=nil { return nil,e }
	return addr,nil
}
func tcpFromUdp(ua net.Addr, n int) (*net.TCPAddr,bool){
	u2,ok := ua.(*net.UDPAddr)
	if !ok { return nil,false }
	return &net.TCPAddr{ u2.IP, n },true
}

var empty Packet

type Node struct{
	Conn net.PacketConn
	Id []byte
	IdEnc string
	Table *ContactTable
	IHCh chan *IHaveMsg
	Bucket bucket.Bucket
	BucketServicePort string
	Debug bool
}
func CreateNode(conn net.PacketConn, id []byte) *Node{
	n := new(Node)
	n.Conn = conn
	n.Id = id
	n.IdEnc = hex.EncodeToString(id)
	n.Table = &ContactTable{Data:make([]Contact,0,1000)}
	n.IHCh = make(chan *IHaveMsg,10)
	return n
}
func (n *Node) Ping(addr net.Addr){
	var pck Packet
	pck.Id=n.IdEnc
	pck.Q="ping"
	bytes,e := json.Marshal(pck)
	if e!=nil { return }
	if n.Debug { fmt.Println("ping",addr) }
	n.Conn.WriteTo(bytes,addr)
}
func (n *Node) Search(id []byte){
	var pck Packet
	pck.Id=n.IdEnc
	pck.Q="search"
	pck.OtherId = hex.EncodeToString(id)
	bytes,e := json.Marshal(pck)
	if e!=nil { return }
	if n.Debug { fmt.Println("search",hex.EncodeToString(id)) }
	closest,_,ok := n.Table.GetClosest(id)
	if ok { n.Conn.WriteTo(bytes,closest.Addr) }
}
func (n *Node) Recommend(addr, other net.Addr){
	var pck Packet
	pck.Id=n.IdEnc
	pck.Q="recommend"
	bytes,e := json.Marshal(pck)
	if e!=nil { return }
	if n.Debug { fmt.Println("recommend",addr,"->",other) }
	n.Conn.WriteTo(bytes,addr)
}
func (n *Node) Exists(hash []byte) bool {
	if n.Bucket==nil { return false }
	return n.Bucket.Exists(hash)
}
func (n *Node) ihave(id string, addr net.Addr) {
	var pck Packet
	pck.Id=n.IdEnc
	pck.Q="ihave"
	pck.OtherId = id
	pck.Appendix = n.BucketServicePort
	bytes,e := json.Marshal(pck)
	if e!=nil { return }
	if n.Debug { fmt.Println("ihave",id) }
	n.Conn.WriteTo(bytes,addr)
}
func (n *Node) Run(){
	var pck Packet
	buf := make([]byte,1024)
	for {
		rt,a,e := n.Conn.ReadFrom(buf)
		if e!=nil { continue }
		if n.Debug { fmt.Println(string(buf[:rt])) }
		pck = empty
		e = json.Unmarshal(buf[:rt],&pck)
		if e!=nil {
			if n.Debug { fmt.Println("json.Unmarshal:",e) }
			continue
		}
		switch pck.Q{
		case "recommend":
			a2,e := appendixToUdpAddress(pck.Appendix)
			if e!=nil { continue }
			n.Ping(a2)
		case "ping":
			enc,e := json.Marshal(Packet{Id:n.IdEnc,Q:"pong",OtherId:pck.Id})
			if e==nil {
				n.Conn.WriteTo(enc,a)
				if n.Debug { fmt.Println("ping",a) }
			}else{
				if n.Debug { fmt.Println("ping failed",a,e) }
			}
			fallthrough
		case "pong":
			d,e := HammingDistanceRH(n.Id,pck.Id)
			if e==nil {
				for a2 := range n.Table.GetCloser(pck.Id,d) {
					n.Recommend(a,a2)
				}
				if ok,_ := n.Table.SearchId(strings.ToLower(pck.Id)) ; !ok {
					n.Table.Insert(Contact{
						Distance:d,
						Id:strings.ToLower(pck.Id),
						Addr:a,
					})
					n.Table.SLimit(10)
				}
			}else{
				if n.Debug { fmt.Println("HammingDistanceRH:",e) }
			}
		case "search":
			bid,e := hex.DecodeString(pck.OtherId)
			mydist := HammingDistance(n.Id,bid)
			if e!=nil || len(bid)!=64 {
				if n.Debug { fmt.Println("search error:",e,"||",len(bid),"!=",64) }
				continue
			}
			if n.Exists(bid) {
				a2,e := appendixToUdpAddress(pck.Appendix)
				if e!=nil { continue }
				if a2!=nil {
					n.ihave(pck.OtherId,a2)
				}else{
					n.ihave(pck.OtherId,a)
				}
			}else{
				if pck.Appendix==nil { pck.Appendix=a.String() }
				bytes,e := json.Marshal(pck)
				if e!=nil { continue }
				closest,dist,ok := n.Table.GetClosest(bid)
				if ok && dist<=mydist { n.Conn.WriteTo(bytes,closest.Addr) } // never send up!
			}
		case "ihave":
			bid,e := hex.DecodeString(pck.OtherId)
			if e!=nil || len(bid)!=64 {
				if n.Debug { fmt.Println("ihave error:",e,"||",len(bid),"!=",64) }
				continue
			}
			ports,ok := pck.Appendix.(string)
			if !ok { continue }
			port,e := strconv.Atoi(ports)
			if e!=nil { continue }
			tcp,ok := tcpFromUdp(a,port)
			if !ok { continue }
			select {
				case n.IHCh <- &IHaveMsg{bid,tcp}:
					if n.Debug { fmt.Println("ihave=",port) }
				default:
			}
		}
	}
}


