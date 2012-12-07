package bucket

import "crypto/sha512"

// Store takes a block and stores it in the Bucket using a
// Sha1-Hash as key, wich is returned
func StoreBlock(b Bucket, block []byte) (hash []byte, err error) {
	h := sha512.New()
	h.Write(block)
	hash = h.Sum([]byte{})
	err = b.Store(hash,block)
	return
}
