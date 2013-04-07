package relaxdht

import . "libblockify/constants"
import "encoding/hex"
import "errors"

var bitcount [256]int

func init() {
	for i:=0 ; i<256 ; i++ {
		j:=i
		k:=0
		for j!=0 { j&=j-1 ; k++ }
		bitcount[i]=k
	}
}

// Calculates the hamming distance between 2 hashes
func HammingDistance(id1, id2 []byte) int {
	j := 0
	for i:=0 ; i<HashSize ; i++ {
		j+=bitcount[id1[i]^id2[i]]
	}
	return j
}

func HammingDistanceRH(id1 []byte, id2s string) (int,error) {
	id2,e := hex.DecodeString(id2s)
	if e!=nil { return 0,e }
	if len(id2)!=64 { return 0,errors.New("size error") }
	return HammingDistance(id1,id2),nil
}
