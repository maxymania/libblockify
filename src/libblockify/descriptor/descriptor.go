package descriptor

import "io"
import "libblockify/blockheader"
import . "libblockify/constants"

func blockCopy(i int, block []byte) (r []byte){
	r = make([]byte,64)
	copy(r,block[i*64:64+(i*64)])
	return
}

type DescriptorBlock struct{
	Header *blockheader.Header
	Tuples [][][]byte
}

func (db *DescriptorBlock) AddBlock(tuple [][]byte) bool{
	if len(tuple)!=db.Header.TupleSize { return false }
	db.Tuples = append(db.Tuples,tuple)
	return true
}
func (db *DescriptorBlock) UpdateHeader() {
	db.Header.NumOfTuples = len(db.Tuples)
}
func (db *DescriptorBlock) Serialize(rnd io.Reader, block []byte) bool{
	ts := db.Header.TupleSize
	nt := db.Header.NumOfTuples
	if (ts*nt)>MaxHashes { return false }
	i := 1
	hd,e := db.Header.Serialize(rnd)
	if e!=nil { return false }
	copy(block[:64],hd)
	for _,tuple := range(db.Tuples) {
		for _,elem := range tuple {
			copy(block[i*64:64+(i*64)],elem)
			i++
		}
	}
	n := i*64
	if n<BlockSize {
		_,e = io.ReadFull(rnd,block[n:])
		if e!=nil { return false }
	}
	return true
}
func (db *DescriptorBlock) Parse(block []byte) bool{
	if !db.Header.Parse(block[:64]) { return false }
	i := 1
	ts := db.Header.TupleSize
	nt := db.Header.NumOfTuples
	db.Tuples = make([][][]byte,nt)
	for ri := range db.Tuples {
		tuple := make([][]byte,ts)
		for ti := range tuple {
			tuple[ti] = blockCopy(i,block)
			i++
		}
		db.Tuples[ri] = tuple
	}
	return true
}
