package bucket

type Bucket interface{
	Store(hash, block []byte) error
	Load(hash []byte) ([]byte,error)
	
	// Efficient Load without block allocation
	ELoad(hash, block []byte) error
	
	// Efficient Testing whether a block to that given hash exists
	Exists(hash []byte) bool
	
	// Lists the Hashes and pushes it in the given channel
	// it shall not close() the channel
	ListUp(hashes chan <- []byte)
}
