package generalutils

import (
	"libblockify/blockutil"
	"libblockify/bucket"
	"io"
)

// block1 should be already initialized
func GenerateTupleFromBlock(block1, block2 []byte, rand io.Reader, bck bucket.Bucket, tupleSize int) (tuple [][]byte,e error) {
	tuple = make([][]byte,tupleSize)
	for i:=1 ; i<tupleSize ; i++ {
		e = blockutil.RandomBlock(rand,block2)
		if e!=nil { return }
		tuple[i],e = bucket.StoreBlock(bck,block2)
		if e!=nil { return }
		blockutil.XorBlock(block1,block2)// block1 ^= block2
	}
	tuple[0],e = bucket.StoreBlock(bck,block1)
	if e!=nil { return }
	return
}

// generates a tuple from a block readed from the src io.Reader.
func GenerateTuple(block1, block2 []byte, rand, src io.Reader, bck bucket.Bucket, tupleSize int) (readed int,tuple [][]byte,e error) {
	tuple = make([][]byte,tupleSize)
	readed = blockutil.FillBlock(rand,src,block1)
	for i:=1 ; i<tupleSize ; i++ {
		e = blockutil.RandomBlock(rand,block2)
		if e!=nil { return }
		tuple[i],e = bucket.StoreBlock(bck,block2)
		if e!=nil { return }
		blockutil.XorBlock(block1,block2)// block1 ^= block2
	}
	tuple[0],e = bucket.StoreBlock(bck,block1)
	if e!=nil { return }
	return
}

func DecodeTuple(block1, block2 []byte, bck bucket.Bucket, tuple [][]byte) (e error) {
	ts := len(tuple)
	e = bck.ELoad(tuple[0],block1)
	if e!=nil { return }
	for i:=1 ; i<ts ; i++ {
		e = bck.ELoad(tuple[i],block2)
		if e!=nil { return }
		blockutil.XorBlock(block1,block2)
	}
	return
}

func TestDecodeTuple(block1, block2 []byte, bck bucket.Bucket, tuple [][]byte, hashes chan []byte) (ok bool) {
	ts := len(tuple)
	ok = true
	e := bck.ELoad(tuple[0],block1)
	if e!=nil { hashes <- tuple[0] ; ok = false }
	for i:=1 ; i<ts ; i++ {
		e2 := bck.ELoad(tuple[i],block2)
		if e2!=nil { hashes <- tuple[i] ; ok = false }
		blockutil.XorBlock(block1,block2)
	}
	return
}
