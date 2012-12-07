package blockheader

import "io"
import "regexp"
import "fmt"
import "errors"
import "strconv"

var parser = regexp.MustCompile(`(ts|nt|ddblock):([tf]|[0-9]+)`)


// The Text Header Structure is "ts:3;nt:983;ddblock:f\0" + Random...
// 
// ts = Tuple Size; nt = Number of Tuples; ddblock = DDBlock = Childs Are Descriptors if true ("t")
type Header struct{
	TupleSize   int  // Sizee of each Tuple
	NumOfTuples int  // Number of Tuples
	DDBlock    bool // Childs Are Descriptors
}
func NewHeader() *Header {
	h := new(Header)
	h.TupleSize = 3 // it should be h.TupleSize >= 3
	return h
}
func (h *Header) Serialize(rnd io.Reader) ([]byte,error) {
	var ddb rune
	if h.DDBlock { ddb = 't' } else { ddb = 'f' }
	s := fmt.Sprintf("ts:%d;nt:%d;ddblock:%c",h.TupleSize,h.NumOfTuples,ddb)
	lens := len(s)
	if lens>64 { return nil,errors.New("Header size error, to big!") }
	if lens<64 {
		r := make([]byte,64)
		copy(r,[]byte(s))
		r[lens]=0
		if lens<63 { io.ReadFull(rnd,r[lens+1:]) }
		return r,nil
	}
	return []byte(s),nil
}
func (h *Header) Parse(bh []byte) bool{
	var e error
	i := 0
	for ; i<64 ; i++ { if bh[i]==0 { break } }
	s := string(bh[:i])
	for _,t := range parser.FindAllStringSubmatch(s,4) {
		switch t[1] {
		case "ts":
			h.TupleSize,e=strconv.Atoi(t[2])
		case "nt":
			h.NumOfTuples,e=strconv.Atoi(t[2])
		case "ddblock":
			h.DDBlock = (t[2]=="t")
		}
		if e!=nil { return false }
	}
	return true
}
