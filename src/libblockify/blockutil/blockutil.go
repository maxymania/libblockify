package blockutil

import . "libblockify/constants"
import "io"

func XorBlock(ack, other []byte) {
	for i := 0 ; i < BlockSize ; i++ {
		ack[i] ^= other[i]
	}
}

// Read a Block from a file (or another Stream)
func FillBlock(rnd, val io.Reader, block []byte) int{
	if len(block)!=BlockSize { return -1 }// Should not happen
	r,_ := io.ReadFull(val,block)
	if r<BlockSize {
		io.ReadFull(rnd,block[r:])
	}
	return r
}

// Randomize The Block/ Create a Randomized Block
func RandomBlock(rnd io.Reader, block []byte) error{
	_,e := io.ReadFull(rnd,block)
	return e
}

// Allocates a Block
func AllocateBlock() []byte{ return make([]byte,BlockSize) }

