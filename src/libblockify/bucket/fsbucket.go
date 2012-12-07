package bucket

import "os"
import "io"
import "io/ioutil"
import "encoding/base32"

type FsBucket string

func (fb FsBucket) Name(hash []byte) string{
	return string(fb)+base32.StdEncoding.EncodeToString(hash)
}
func (fb FsBucket) Store(hash, block []byte) error {
	return ioutil.WriteFile(fb.Name(hash),block,0666)
}
func (fb FsBucket) Load(hash []byte) ([]byte,error) {
	return ioutil.ReadFile(fb.Name(hash))
}

func (fb FsBucket) ELoad(hash, block []byte) error {
	f,e := os.Open(fb.Name(hash))
	if e!=nil { return e }
	defer f.Close()
	_,e = io.ReadFull(f,block)
	return e
}
