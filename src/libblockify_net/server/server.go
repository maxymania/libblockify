package server

import (
	"regexp"
	"io"
	"net"
	"net/textproto"
	// "libblockify/blockutil"
	"libblockify/bucket"
	. "libblockify/constants"
	// "errors"
	"encoding/base64"
	// "fmt"
)

var pullblock *regexp.Regexp

func init() {
	pullblock = regexp.MustCompile(`^PULLBLOCK (.+)$`)
}

type Connection struct{
	conn *textproto.Conn
	bck bucket.Bucket
	
	// TODO: consider a more efficient way to allocate 2 blocks
	block1raw [BlockSize]byte
	block2raw [BlockSize]byte
	
	block1 []byte
	block2 []byte
}

func Connect(conn io.ReadWriteCloser,bck bucket.Bucket) *Connection{
	c := new(Connection)
	c.conn=textproto.NewConn(conn)
	c.bck=bck
	c.block1=c.block1raw[:]
	c.block2=c.block2raw[:]
	return c
}

func Serve(listener net.Listener,bck bucket.Bucket){
	for {
		c,e := listener.Accept()
		if e!=nil { continue }
		conn := Connect(c,bck)
		go conn.Serve()
	}
}

func (c *Connection) donothing(id uint) {
	c.conn.StartResponse(id)
	c.conn.EndResponse(id)
}

func (c *Connection) sendLine(id uint, code int, msg string){
	c.conn.StartResponse(id)
	defer c.conn.EndResponse(id)
	c.conn.PrintfLine("%d %s",code,msg)
}

func (c *Connection) sendBlock(id uint, key []byte) {
	c.conn.StartResponse(id)
	defer c.conn.EndResponse(id)
	e := c.bck.ELoad(key,c.block2)
	if e==nil {
		c.conn.PrintfLine("100 OK")
		c.conn.W.Write(c.block2)
		c.conn.W.Flush()
	} else {
		c.conn.PrintfLine("904 NOT FOUND")
	}
}

func (c *Connection) Serve(){
	for {
		r := c.conn.Next()
		c.conn.StartRequest(r)
		s,e := c.conn.ReadLine()
		if e!=nil {break}
		pbmatch := pullblock.FindStringSubmatch(s)
		switch {
		case s=="PUSHBLOCK":
			_,e = io.ReadFull(c.conn.R,c.block1)
			if e==nil{
				bucket.StoreBlock(c.bck,c.block1)
			}
			c.conn.EndRequest(r)
			go c.donothing(r)
		case pbmatch!=nil:
			key,e := base64.URLEncoding.DecodeString(pbmatch[1])
			c.conn.EndRequest(r)
			if e!=nil {
				go c.sendLine(r,901,"base64 error")
			} else {
				go c.sendBlock(r,key)
			}
		default:
			c.conn.EndRequest(r)
			go c.sendLine(r,909,"command error")
		}
	}
	// println("close Connection")
	c.conn.Close()
}

