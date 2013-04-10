package f2ftp

import (
	"fmt"
	"net"
	"net/textproto"
	"io"
	"bufio"
	"bytes"
	"regexp"
	. "libblockify/constants"
	"libblockify/bucket"
	"crypto/sha512"
	"encoding/base64"
)

var rxreq = regexp.MustCompile(`^([A-Z]+) (.*)$`)

type BlockAcceptor interface{
	Accept(hash []byte) bool
}
type BoolBlockAcceptor bool
func (b BoolBlockAcceptor)Accept(hash []byte) bool { return bool(b) }

const NullBlockAcceptor BoolBlockAcceptor = false
const AllBlockAcceptor BoolBlockAcceptor = true

type Service struct{
	acc BlockAcceptor
	bck bucket.Bucket
	
	writerInterrupt chan int
	writerReturn chan int
	
	rpull chan []byte
	
	rblock []byte
	wblock []byte
	
	// Quit Requests made by the user
	ChQuit chan int
	// Pull Requests made by the user
	ChPull chan []byte
	Debug bool
}

func NewService(acc BlockAcceptor, bck bucket.Bucket) *Service {
	s := &Service{
		acc:acc,
		bck:bck,
		writerInterrupt: make(chan int,1),
		writerReturn: make(chan int,1),
		rpull:make(chan []byte,128),
		rblock:make([]byte,BlockSize),
		wblock:make([]byte,BlockSize),
		ChQuit:make(chan int,1),
		ChPull:make(chan []byte,128),
	}
	return s
}

func (s *Service) reader(r *textproto.Reader) {
	for {
		str,e := r.ReadLine()
		if e!=nil { break }
		req := rxreq.FindStringSubmatch(str)
		if req==nil { break }
		if req[1]=="QUIT" { break }
		switch req[1] {
			case "PULL":
				if s.Debug { fmt.Println("got pull",req[2]) }
				k,e := base64.StdEncoding.DecodeString(req[2])
				if e==nil { s.rpull <- k }
			case "PUSH":
				_,e = io.ReadFull(r.R,s.rblock)
				if e==nil {
					h := sha512.New()
					h.Write(s.rblock)
					hash := h.Sum([]byte{})
					if s.acc.Accept(hash) {
						if s.Debug { fmt.Println("got push",base64.StdEncoding.EncodeToString(hash),"accept") }
						s.bck.Store(hash,s.rblock)
					}else{
						if s.Debug { fmt.Println("got push",base64.StdEncoding.EncodeToString(hash),"refuse") }
					}
				}
		}
	}
}

func (s *Service) writer(w io.Writer){
	var buf bytes.Buffer
	for {
		buf.Truncate(0)
		select {
			case k := <- s.rpull:
				if s.bck.ELoad(k,s.wblock)==nil {
					buf.WriteString("PUSH x\r\n")
					buf.Write(s.wblock)
				}
			case <- s.writerInterrupt: goto exitfunc
			case k := <- s.ChPull:
				str := base64.StdEncoding.EncodeToString(k)
				buf.WriteString("PULL "+str+"\r\n")
			case <- s.ChQuit:
				w.Write([]byte("QUIT x\r\n"))
				goto exitfunc
		}
		buf.WriteTo(w)
	}
	exitfunc:
	s.writerReturn <- 1
}

func (s *Service) Serve(conn net.Conn){
	r := textproto.NewReader(bufio.NewReaderSize(conn,128))
	go s.writer(conn)
	s.reader(r)
	s.writerInterrupt <- 1
	<- s.writerReturn
	conn.Close()
}

