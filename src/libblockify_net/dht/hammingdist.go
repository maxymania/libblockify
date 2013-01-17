package dht

import . "libblockify/constants"

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
