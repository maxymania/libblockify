Blockify
========

A Filesharing System similar to OFF (Owner Free Filesystem)

libblockify
-----------

libblockify is the reference Implementation of Blockify. And it is a Library (thats why its called "lib")

Status
------

There is not a working program, just a library.
Insertion and extraction already works.
The Network Component is under heavy developement and not stable.

here is that code i used t "test" the functionality(file pathes changed)

```go
package main
import . "libblockify/constants"
import "crypto/rand"
import "libblockify/blockheader"
import "libblockify/descriptor"
import "libblockify/blockutil"
import "libblockify/bucket"
import "libblockify/generalutils"
import "libblockify/link"
import "fmt"
import "os"
import "encoding/base32"
// import "strconv"
import "io/ioutil"

const bktpath = `/path/to/bucket/`
const lnkfile = `/path/to/link.txt`
const myfile = `/path/to/song.mp3`
const mynewfile = `/path/to/mysong.mp3`
const mypath = `/path/to/_dest/`

var mybucket = bucket.FsBucket(bktpath)

func encode2(){
	src,e := os.Open(myfile)
	if e!=nil { fmt.Println(e) ; return }
	defer src.Close()
	block1 := blockutil.AllocateBlock()
	block2 := blockutil.AllocateBlock()
	// rest,tuple,e := generalutils.UploadStreamSimple(block1,block2,rand.Reader,src,mybucket,3)
	blocked,rest,tuple,e := generalutils.UploadStream(block1,block2,rand.Reader,src,mybucket,3,3)
	fmt.Println("blocked=",blocked)
	url := link.MakeURL(tuple,rest,"song.mp3")
	fmt.Println(url)
	lnk,_ := os.Create(lnkfile)
	fmt.Fprint(lnk,url)
	lnk.Close()
	
	// dest, _:= os.Create(mynewfile)
	// defer dest.Close()
	// generalutils.DownloadStream(block1,block2,dest,mybucket,tuple,rest)
}

func decode2(){
	urlb,_ := ioutil.ReadFile(lnkfile)
	url := string(urlb)
	// fmt.Println(url)
	// fmt.Println(link.ParseURL(url))
	tuple,rest,file,e := link.ParseURL(url)
	if e!=nil { fmt.Println(e) ; return }
	dest, e:= os.Create(mypath+file)
	if e!=nil { fmt.Println(e) ; return }
	defer dest.Close()
	block1 := blockutil.AllocateBlock()
	block2 := blockutil.AllocateBlock()
	generalutils.DownloadStream(block1,block2,dest,mybucket,tuple,rest)
}

func encode(){
	src,e := os.Open(myfile)
	if e!=nil { fmt.Println(e) ; return }
	defer src.Close()
	
	block := blockutil.AllocateBlock()
	block2 := blockutil.AllocateBlock()
	
	hdr := blockheader.NewHeader()
	descr := &descriptor.DescriptorBlock{Header:hdr}
	ts := hdr.TupleSize
	
	for {
		readed,tuple,e := generalutils.GenerateTuple(block,block2,rand.Reader,src,mybucket,ts)
		if e!=nil { fmt.Println("generalutils.GenerateTuple",e) ; return }
		descr.AddBlock(tuple)
		if readed<BlockSize { break }
	}
	
	descr.UpdateHeader()
	if !descr.Serialize(rand.Reader,block) {
		fmt.Println("failed")
	}
	
	hash,e := bucket.StoreBlock(mybucket,block)
	if e!=nil { fmt.Println(e) ; return }
	lnk,_ := os.Create(lnkfile)
	fmt.Fprint(lnk,base32.StdEncoding.EncodeToString(hash))
	lnk.Close()
	fmt.Println("done with it")
}

func decode(){
	eh,_ := ioutil.ReadFile(lnkfile)
	h,_ := base32.StdEncoding.DecodeString(string(eh))
	
	block := blockutil.AllocateBlock()
	block2 := blockutil.AllocateBlock()
	
	mybucket.ELoad(h,block)
	descr := &descriptor.DescriptorBlock{Header:new(blockheader.Header)}
	descr.Parse(block)
	// ts := descr.Header.TupleSize
	// fmt.Println(descr.Header)
	
	dest, _:= os.Create(mynewfile)
	defer dest.Close()
	for _,tuple := range descr.Tuples {
		generalutils.DecodeTuple(block,block2,mybucket,tuple)
		dest.Write(block)
	}
}

func main(){
	// encode()
	// decode()
	// encode2()
	decode2()
	fmt.Println("hallo welt!")
}
```

License
-------

This software is licensed under the MIT-License. See LICENSE file.
