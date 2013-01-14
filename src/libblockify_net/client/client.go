package client

import (
	"io"
	// "net"
	"net/textproto"
	// "libblockify/blockutil"
	// "libblockify/bucket"
	. "libblockify/constants"
	"errors"
	"encoding/base64"
)

type Connection struct{
	conn *textproto.Conn
}

func Dial(network, addr string) (*Connection,error){
	c,e := textproto.Dial(network,addr)
	return &Connection{c},e
}

func (c *Connection) Push(block []byte) error{
	if len(block)!=BlockSize { return errors.New("Wrong block size") }
	r := c.conn.Next()
	c.conn.StartRequest(r)
	defer func(){
		c.conn.EndRequest(r)
		go func(){
			c.conn.StartResponse(r)
			c.conn.EndResponse(r)
		}()
	}()
	e := c.conn.PrintfLine("PUSHBLOCK")
	if e!=nil { return e }
	_,e = c.conn.W.Write(block)
	if e!=nil { return e }
	e = c.conn.W.Flush()
	return e
}

func (c *Connection) Pull(key []byte,block []byte) error{
	if len(block)!=BlockSize { return errors.New("Wrong block size") }
	r,e := c.conn.Cmd("PULLBLOCK %s",base64.URLEncoding.EncodeToString(key))
	if e!=nil { return e }
	c.conn.StartResponse(r)
	defer c.conn.EndResponse(r)
	_,_,e = c.conn.ReadCodeLine(100)
	if e!=nil { return e }
	_,e = io.ReadFull(c.conn.R,block)
	return e
}

func (c *Connection) Close() error { return c.conn.Close() }
