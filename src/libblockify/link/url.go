package link

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"encoding/base64"
	"errors"
)

var decodeurl *regexp.Regexp
var parseError = errors.New("parse error")

func init() {
	decodeurl = regexp.MustCompile(`^blockify\://(.+?)//(\d+)/(.+)`)
}

func MakeURL(tuple [][]byte, rest int, filename string) string{
	var url bytes.Buffer
	url.WriteString("blockify://")
	for _,item := range tuple {
		url.WriteString(base64.URLEncoding.EncodeToString(item))
		url.WriteRune('/')
	}
	url.WriteRune('/')
	url.WriteString(strconv.Itoa(rest))
	url.WriteRune('/')
	url.WriteString(filename)
	return url.String()
}

func ParseURL(url string) (tuple [][]byte, rest int, filename string, err error) {
	data := decodeurl.FindStringSubmatch(url)
	if data==nil {
		err=parseError
		return
	}
	tpl := strings.Split(data[1],"/")
	tuple = make([][]byte,len(tpl))
	for i,s := range tpl {
		tuple[i],err = base64.URLEncoding.DecodeString(s)
		if err!=nil { return }
	}
	rest,err = strconv.Atoi(data[2])
	filename = data[3]
	return
}
