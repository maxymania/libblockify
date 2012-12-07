package bucket

type Bucket interface{
	Store(hash, block []byte) error
	Load(hash []byte) ([]byte,error)
	
	// Efficient Load without block allocation
	ELoad(hash, block []byte) error
}
