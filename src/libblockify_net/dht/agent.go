package dht

import "net"
import "strconv" // ParseInt
import "libblockify/bucket"
import . "libblockify/constants"
import "encoding/base64"
import "errors"

var hashSizeError = errors.New("hash size error")

type AddrConv func(str string) (net.Addr, error)
func AddrConvUDP (str string) (net.Addr, error) {
	return net.ResolveUDPAddr("udp",str)
}

func IdEncode(id []byte) string { return base64.URLEncoding.EncodeToString(id) }
func IdDecode(id string) ([]byte,error) {
	id2,e := base64.URLEncoding.DecodeString(id)
	if e==nil && len(id2)!=HashSize { e=hashSizeError }
	return id2,e
}

type CmdRequest struct{
	Cmd Cmd
	Id []byte
	Addr net.Addr
}

type AgentThread struct{
	OwnNodeID []byte
	MainList Nlist
	NearList Nlist
	Agent Agent
	AddrConv AddrConv
	Bucket bucket.Bucket
	MaxTTL int64
	requestCh chan CmdRequest
	requestCh2 chan CmdRequest
	propagateCh chan Cmd
	pongCh chan Cmd
}
func (at *AgentThread) Constructor() {
	at.requestCh  = make(chan CmdRequest,128)
	at.requestCh2 = make(chan CmdRequest,128)
	at.propagateCh = make(chan Cmd,128)
	at.pongCh = make(chan Cmd,128)
}
func (at *AgentThread) dispatcher(){
	for cmd := range at.Agent.ReaderCh {
		switch cmd.CmdId {
		case Request: {
			k,e := IdDecode(cmd.CmdPar[2])
			if e!=nil { continue }
			a := cmd.Addr
			if len(cmd.CmdPar)==4 {
				a2,e := at.AddrConv(cmd.CmdPar[3])
				if e!=nil { continue }
				a=a2
			}
			cmd2 := CmdRequest{cmd,k,a}
			at.requestCh  <- cmd2
			at.requestCh2 <- cmd2
			}
		case Propagate:
			at.propagateCh <- cmd
		case Ping:
			at.Agent.Pong(cmd.Addr,IdEncode(at.OwnNodeID))
		
		}
	}
}
func (at *AgentThread) requestDisp() {
	for cmd := range at.requestCh {
		a := cmd.Addr
		if at.Bucket!=nil && at.Bucket.Exists(cmd.Id) {
			at.Agent.IHave(a,cmd.Cmd.CmdPar[2])
		}
	}
}
func (at *AgentThread) requestDisp2() {
	for cmd := range at.requestCh2 {
		ttl,e := strconv.ParseInt(cmd.Cmd.CmdPar[1],10,64)
		if e!=nil { continue }
		if ttl>at.MaxTTL { ttl = at.MaxTTL }
		ttl-=1
		dist := HammingDistance(at.OwnNodeID,cmd.Id)
		nc := make(chan *Node)
		go func(){
			for n := range nc {
				dist2 := HammingDistance(n.Id,cmd.Id)
				if dist2<dist {
					at.Agent.Request(n.Addr,ttl,cmd.Cmd.CmdPar[2],cmd.Addr)
				}
			}
		}()
		// go func(){
			at.MainList.Iterate(nc)
			close(nc)
		// }()
	}
}
func (at *AgentThread) propagateDisp() {
	for cmd := range at.propagateCh {
		k,e := IdDecode(cmd.CmdPar[1])
		if e!=nil { continue }
		a,e := at.AddrConv(cmd.CmdPar[2])
		if e!=nil { continue }
		n := &Node{k,a}
		at.MainList.InsertNode(n)
		at.NearList.InsertNode(n)
	}
}
func (at *AgentThread) pongDisp() {
	for cmd := range at.pongCh {
		k,e := IdDecode(cmd.CmdPar[1])
		if e!=nil { continue }
		a := cmd.Addr
		n := &Node{k,a}
		at.MainList.InsertNode(n)
		at.NearList.InsertNode(n)
	}
}
// func (at *AgentThread) 

func (at *AgentThread) Start() {
	go at.dispatcher()
	go at.requestDisp()
	go at.requestDisp2()
	go at.propagateDisp()
	go at.pongDisp()
}

