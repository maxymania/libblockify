package generalutils

import . "libblockify/constants"

import (
	"libblockify/blockheader"
	"libblockify/descriptor"
	"libblockify/bucket"
	"io"
	"errors"
)

// This is a simple update function. It doesnt create ddblocks (Descriptors of Descriptors).
func UploadStreamSimple(block1, block2 []byte, rand, src io.Reader, bck bucket.Bucket, tupleSize int) (rest int, tuple [][]byte,e error) {
	hdr := blockheader.NewHeader()
	descr := &descriptor.DescriptorBlock{Header:hdr}
	hdr.TupleSize = tupleSize
	i := 0
	for i<MaxHashes {
		readed,tuple,e := GenerateTuple(block1, block2, rand, src, bck, tupleSize)
		if e!=nil { return 0,nil,e }
		descr.AddBlock(tuple)
		i += tupleSize
		rest = readed
		if readed<BlockSize { break }
	}
	descr.UpdateHeader()
	if !descr.Serialize(rand,block1) {
		return 0,nil,errors.New("serialisation-error")
	}
	tuple,e = GenerateTupleFromBlock(block1,block2,rand,bck,tupleSize)
	return
}

func DownloadStreamSimple(block1, block2 []byte, dest io.Writer, bck bucket.Bucket, tuple [][]byte, rest int) (e error){
	if rest==0 { rest = BlockSize }
	e = DecodeTuple(block1,block2,bck,tuple)
	if e!=nil { return }
	descr := &descriptor.DescriptorBlock{Header:new(blockheader.Header)}
	descr.Parse(block1)
	last := len(descr.Tuples)-1
	for i,tuple := range descr.Tuples {
		DecodeTuple(block1,block2,bck,tuple)
		if i==last {
			_,e = dest.Write(block1[:rest])
		} else {
			_,e = dest.Write(block1)
		}
		if e!=nil { return }
	}
	return
}

func UploadStream(block1, block2 []byte, rand, src io.Reader, bck bucket.Bucket, tupleSize int, depth int) (broken bool,rest int, tuple [][]byte,e error) {
	broken = false
	hdr := blockheader.NewHeader()
	descr := &descriptor.DescriptorBlock{Header:hdr}
	hdr.TupleSize = tupleSize
	i := 0
	if depth<=1 {
		for i<MaxHashes {
			readed,tuple,e := GenerateTuple(block1, block2, rand, src, bck, tupleSize)
			if e!=nil { return false,0,nil,e }
			descr.AddBlock(tuple)
			i += tupleSize
			rest = readed
			if readed<BlockSize { broken = true ; break }
		}
	} else {
		for i<MaxHashes {
			broken2,rest2,tuple2,e2 := UploadStream(block1,block2,rand,src,bck,tupleSize,depth-1)
			if e2!=nil { return false,0,nil,e2 }
			descr.AddBlock(tuple2)
			i += tupleSize
			rest = rest2
			if broken2 || rest2<BlockSize { broken = true ; break }
		}
	}
	if len(descr.Tuples) == 1 {
		tuple = descr.Tuples[0]
		return
	}
	descr.UpdateHeader()
	if !descr.Serialize(rand,block1) {
		return broken,0,nil,errors.New("serialisation-error")
	}
	tuple,e = GenerateTupleFromBlock(block1,block2,rand,bck,tupleSize)
	return
}

func DownloadStream(block1, block2 []byte, dest io.Writer, bck bucket.Bucket, tuple [][]byte, rest int) (e error){
	if rest==0 { rest = BlockSize }
	e = DecodeTuple(block1,block2,bck,tuple)
	if e!=nil { return }
	descr := &descriptor.DescriptorBlock{Header:new(blockheader.Header)}
	descr.Parse(block1)
	last := len(descr.Tuples)-1
	if descr.Header.DDBlock {
		for i,tuple2 := range descr.Tuples {
			mr := BlockSize
			if i==last { mr = rest }
			e = DownloadStream(block1,block2,dest,bck,tuple2,mr)
			if e!=nil { return }
		}
	} else {
		for i,tuple2 := range descr.Tuples {
			DecodeTuple(block1,block2,bck,tuple2)
			if i==last {
				_,e = dest.Write(block1[:rest])
			} else {
				_,e = dest.Write(block1)
			}
			if e!=nil { return }
		}
	}
	return
}