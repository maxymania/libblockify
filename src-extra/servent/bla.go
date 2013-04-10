package main

import (
	"fmt"
	"os"
	"encoding/json"
	"encoding/hex"
	"crypto/sha512"
	"net"
	"net/http"
	"time"
)

import "libblockify/bucket"

import "libblockify_net/glue"

import "webgui"

/*
const bktpath = `K:\developement\golang\blockify_tools\experiments\_bucket\`
var mybucket = bucket.FsBucket(bktpath)
*/

type Config struct{
	Bucket string `json:"bucket"`
	Ip string `json:"ip,omitempty"`
	DhtPort string `json:"dht-port"`
	BucketPort string `json:"bucket-port"`
	Gen string `json:"gen"`
	Http []string `json:"http"`
}

func hash(gen []byte) []byte{
	h := sha512.New()
	h.Write(gen)
	return h.Sum([]byte{})
}

func main(){
	var addrs []string
	var config Config
	{
		f,e := os.Open(os.Args[1])
		fmt.Println("config:",os.Args[1])
		if e!=nil { fmt.Println(e) ; return }
		e = json.NewDecoder(f).Decode(&config)
		if e!=nil { fmt.Println(e) ; return }
		f.Close()
	}
	mybucket := bucket.FsBucket(config.Bucket)
	{
		f,e := os.Open(os.Args[2])
		fmt.Println("addrs:",os.Args[2])
		if e!=nil { fmt.Println(e) ; return }
		e = json.NewDecoder(f).Decode(&addrs)
		if e!=nil { fmt.Println(e) ; return }
		f.Close()
	}
	fmt.Println(config)
	gen,e := hex.DecodeString(config.Gen)
	if e!=nil { fmt.Println(e) ; return }
	
	id := hash(gen)
	
	ht,e := net.Listen(config.Http[0],config.Http[1])
	if e!=nil { fmt.Println(e) ; return }
	
	g := new(glue.Glue).Init(id,mybucket,config.Ip,config.DhtPort,config.BucketPort)
	g.DebugOn()
	go g.ServeBucket()
	go g.ServeDht()
	go g.RunPull()
	
	hashes := make(chan []byte)
	go func(){
		for h := range hashes {
			g.Want(h)
		}
	}()
	h := webgui.NewHandler(g.GetBucket(),hashes)
	go func(){
		time.Sleep(time.Second)
		for _,a := range addrs {
			time.Sleep(time.Millisecond*100)
			g.PingUdp(a)
		}
	}()
	http.Serve(ht,h)
}
