package glue

import "net"
import "sync"
import "time"
import "encoding/hex"
import "libblockify/bucket"
import "libblockify_net/relaxdht"
import "libblockify_net/f2ftp"

type acceptor struct{
	m sync.Mutex
	wantMap map[string]bool
}
func (a *acceptor) want(hash []byte){
	a.m.Lock(); defer a.m.Unlock()
	id := hex.EncodeToString(hash)
	a.wantMap[id]=true
}
func (a *acceptor) unwant(hash []byte){
	a.m.Lock(); defer a.m.Unlock()
	id := hex.EncodeToString(hash)
	_,ok := a.wantMap[id]
	if ok { delete(a.wantMap,id) }
}
func (a *acceptor) Accept(hash []byte) bool{
	a.m.Lock(); defer a.m.Unlock()
	id := hex.EncodeToString(hash)
	ok1,ok2 := a.wantMap[id]
	return ok1&&ok2
}


type hBucket struct{
	Acc *acceptor
	Bck bucket.Bucket
	Dht *relaxdht.Node
}
func (hb *hBucket) Store(hash, block []byte) (e error){
	e = hb.Bck.Store(hash,block)
	hb.Dht.Push(hash)
	if e==nil { hb.Acc.unwant(hash) }
	return
}
func (hb *hBucket) Load(hash []byte) ([]byte,error) { return hb.Bck.Load(hash) }
func (hb *hBucket) ELoad(hash, block []byte) error { return hb.Bck.ELoad(hash, block) }
func (hb *hBucket) Exists(hash []byte) bool { return hb.Bck.Exists(hash) }
func (hb *hBucket) ListUp(hashes chan <- []byte) { hb.Bck.ListUp(hashes) }

type Glue struct {
	dht   *relaxdht.Node
	bcks  net.Listener
	// serv f2ftp.Service
	acc   *acceptor
	bck   *hBucket
}
func (g *Glue) Init(id []byte, mybuck bucket.Bucket,ip,dht,bck string) *Glue{
	dhtc,e := net.ListenPacket("udp",ip+":"+dht)
	if e!=nil { panic(e) }
	g.bcks,e = net.Listen("tcp",ip+":"+bck)
	if e!=nil { panic(e) }
	g.acc = &acceptor{wantMap:make(map[string]bool)}
	g.dht = relaxdht.CreateNode(dhtc,id)
	g.bck = &hBucket{g.acc,mybuck,g.dht}
	return g
}
func (g *Glue) GetBucket() bucket.Bucket{ return g.bck }
func (g *Glue) ServeBucket() {
	for{
		conn, err := g.bcks.Accept()
		if err != nil { continue }
		bs := f2ftp.NewService(g.acc,g.bck)
		go bs.Serve(conn)
	}
}
func (g *Glue) ServeDht() {
	g.dht.Run()
}
func (g *Glue) pull(ih *relaxdht.IHaveMsg){
	conn, err := net.DialTCP("tcp",nil,ih.Addr)
	if err != nil { return }
	bs := f2ftp.NewService(g.acc,g.bck)
	go bs.Serve(conn)
	bs.ChPull <- ih.Id
	time.Sleep(time.Second*2)
	bs.ChQuit <- 1
}
func (g *Glue) RunPull(){
	for ih := range g.dht.IHCh{
		go g.pull(ih)
	}
}

func (g *Glue) Want(hash []byte) {
	g.acc.want(hash)
}
func (g *Glue) Ping(addr net.Addr) { g.dht.Ping(addr) }
func (g *Glue) PingUdp(adrs string) {
	addr,e := net.ResolveUDPAddr("udp",adrs)
	if e==nil {
		g.dht.Ping(addr)
	}
}
