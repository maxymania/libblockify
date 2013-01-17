package dht

import "regexp"
import "net"
import "fmt"
import "bytes"
import "sync"

// PROPAGATE <id> <ip:port>
var   rpropagate = regexp.MustCompile(`^PROPAGATE ([^ ]+) ([^ ]+)$`)
const wpropagate = `PROPAGATE %s %s`

// REQUEST <ttl> <id>
var   rrequest = regexp.MustCompile(`^REQUEST ([^0-9]+) ([^ ]+)$`)
const wrequest = `REQUEST %d %s`

// REQUEST <ttl> <id> <ip:port>
var   rrequest2 = regexp.MustCompile(`^REQUEST ([^0-9]+) ([^ ]+) ([^ ]+)$`)
const wrequest2 = `REQUEST %d %s %s`

// I-AM <id>
var   riam = regexp.MustCompile(`^I-AM ([^ ]+)$`)
const wiam = `I-AM %s`

// I-HAVE <id>
var   rihave = regexp.MustCompile(`^I-HAVE ([^ ]+)$`)
const wihave = `I-HAVE %s`

// CALL
const call = `CALL`
var bcall = []byte(call)

// RECALL (ACTIVE|<ip:port>)
var   rrecall = regexp.MustCompile(`^RECALL (ACTIVE|([^ ]+))$`)
const wrecall = `RECALL %s`

// PING <id>
var   rping = regexp.MustCompile(`^PING ([^ ]+)$`)
const wping = `PING %s`

// PONG <id>
var   rpong = regexp.MustCompile(`^PONG ([^ ]+)$`)
const wpong = `PONG %s`

const (
	Propagate = iota
	Request
	IAm
	IHave
	Call
	Recall
	Ping
	Pong
)

type Cmd struct{
	CmdId   int
	CmdPar  []string
	Addr    net.Addr
}

type Agent struct{
	Socket   net.PacketConn
	Open     bool
	ReaderCh chan Cmd
	rbuf     [1024]byte
	wbur     bytes.Buffer
	wlock    sync.Mutex
}
func (a *Agent) Reader() {
	for a.Open {
		a.read()
	}
	close(a.ReaderCh)
}
func (a *Agent) read() {
	buf := a.rbuf[:]
	l,addr,e := a.Socket.ReadFrom(buf)
	if e!=nil { return }
	str := string(buf[:l])
	if pars := rpropagate.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Propagate,pars,addr} ; return
	}
	if pars := rrequest.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Request,pars,addr} ; return
	}
	if pars := rrequest2.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Request,pars,addr} ; return
	}
	if pars := riam.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{IAm,pars,addr} ; return
	}
	if pars := rihave.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{IHave,pars,addr} ; return
	}
	// if pars := rcall.FindStringSubmatch(str); pars!=nil 
	if call==str {
		a.ReaderCh <- Cmd{Call,[]string{},addr} ; return
	}
	if pars := rrecall.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Recall,pars,addr} ; return
	}
	if pars := rping.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Ping,pars,addr} ; return
	}
	if pars := rpong.FindStringSubmatch(str); pars!=nil {
		a.ReaderCh <- Cmd{Pong,pars,addr} ; return
	}
}

func (a *Agent) Propagate(dest net.Addr,id string, addr net.Addr) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	fmt.Fprintf(&a.wbur,wpropagate,id,addr.String())
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) Request(dest net.Addr, ttl int64,id string, addr net.Addr) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	if addr==nil{
		fmt.Fprintf(&a.wbur,wrequest,ttl,id)
	} else {
		fmt.Fprintf(&a.wbur,wrequest2,ttl,id,addr.String())
	}
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) IAm(dest net.Addr, id string) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	fmt.Fprintf(&a.wbur,wiam,id)
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) IHave(dest net.Addr, id string) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	fmt.Fprintf(&a.wbur,wihave,id)
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) Call(dest net.Addr) {
	a.Socket.WriteTo(bcall,dest)
}
func (a *Agent) Recall(dest net.Addr,passive net.Addr) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	id := "ACTIVE"
	if passive!=nil { id = passive.String() }
	fmt.Fprintf(&a.wbur,wrecall,id)
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) Ping(dest net.Addr,id string) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	fmt.Fprintf(&a.wbur,wping,id)
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}
func (a *Agent) Pong(dest net.Addr,id string) {
	a.wlock.Lock()
	defer a.wlock.Unlock()
	fmt.Fprintf(&a.wbur,wpong,id)
	a.Socket.WriteTo(a.wbur.Bytes(),dest)
	a.wbur.Truncate(0)
}

