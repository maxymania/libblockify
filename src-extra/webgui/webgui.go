package webgui

import "os"
import "fmt"
import "text/template"
import "net/http"
import "crypto/rand"
import "path/filepath"
import urllib "net/url"

import "sync"

import "libblockify/generalutils"
import "libblockify/bucket"
import "libblockify/blockutil"
import "libblockify/link"

const mainview = `<!DOCTYPE html>
<html>
<head><title>Blockify</title></head>
<body>
<h1>Blockify</h1>
<a href="/">refresh!</a>
<ul>
{{range $url, $info := .InUse}}
<li>{{ $info.Short | html}} Status={{ $info.Status | html }} <a href="/test?{{ $url | urlquery | html}}">test</a> <a href="/download/{{ $info.Short | urlquery | html }}?{{ $url | urlquery | html}}">download</a> </li>
{{end}}
</ul>
<form action="/add" method="post">
<p>Add urls</p>
<input name="url" type="text" size="100" maxlength="2000">
<input type="submit" value="Add">
</form>
<form action="/upload" method="post">
<p>Upload files</p>
<input name="path" type="text" size="30" maxlength="2000">
<input type="submit" value="Upload">
</form>
</body>
</html>
`

const uploaded = `<!DOCTYPE html>
<html>
<head><title>Blockify Upload</title></head>
<body>
<h1>Blockify</h1>
<a href="/">home</a>
<p>There was the following error:{{ .Error | html }}</p>
<p>here is the url. Store it savely!</p>
<p>{{ .Url | html }}</p>
</body>
</html>
`

type upload struct{
	Error string
	Url string
}

const download = "/download/"

var tmainview = template.Must(template.New("tmainview").Parse(mainview))
var tuploaded = template.Must(template.New("tuploaded").Parse(uploaded))

const (
	None = iota
	Uncomplete
	Testing
	Complete
)

type decodedUrl struct{
	tuple [][]byte
	rest int
	name string
}
type downloadStatus struct{
	decodedUrl
	status int
}
func (ds *downloadStatus) Short() string{ return ds.name }
func (ds *downloadStatus) Status() string{
	switch ds.status{
	case Uncomplete:
		return "uncomplete"
	case Complete:
		return "complete"
	case Testing:
		return "testing"
	}
	return "none"
}

type handler struct{
	m sync.RWMutex
	B bucket.Bucket
	Wants chan []byte
	InUse map[string]*downloadStatus
}
func NewHandler(b bucket.Bucket, hashes chan []byte) http.Handler{
	h := new(handler)
	h.B=b
	h.Wants=hashes
	h.InUse = make(map[string]*downloadStatus)
	return h
}
func (h *handler) testUrl(dls *downloadStatus) {
	block1 := blockutil.AllocateBlock()
    block2 := blockutil.AllocateBlock()
	if generalutils.TestDownloadStream(block1,block2,h.B,dls.tuple,h.Wants) {
		dls.status=Complete
	}else{
		dls.status=Uncomplete
	}
}
func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	switch{
	case path=="/":
		h.m.RLock(); defer h.m.RUnlock()
		tmainview.Execute(resp,h)
		return
	case path=="/add":
		// u := req.URL.RawQuery
		if req.ParseForm()!=nil { goto redirect }
		u := req.Form.Get("url")
		tuple,rest,name,e := link.ParseURL(u)
		if e!=nil { goto redirect }
		h.m.Lock(); defer h.m.Unlock()
		_,ok := h.InUse[u] // do not overwrite anything!
		if !ok {
			h.InUse[u]=&downloadStatus{decodedUrl{tuple,rest,name},None}
		}
		goto redirect 
	case path=="/test":
		u,e := urllib.QueryUnescape(req.URL.RawQuery)
		if e!=nil { goto redirect }
		h.m.Lock(); defer h.m.Unlock()
		obj,ok := h.InUse[u]
		if ok && obj.status!=Testing {
			obj.status=Testing
			go h.testUrl(obj)
		}
		goto redirect
	case len(path)>=len(download) && path[:len(download)]==download:
		u,e := urllib.QueryUnescape(req.URL.RawQuery)
		if e!=nil { goto errresp }
		h.m.Lock()
		obj,ok := h.InUse[u]
		h.m.Unlock()
		if !ok { goto errresp }
		// resp.Header().Add("Content-Type","application/octet-stream")
		block1 := blockutil.AllocateBlock()
		block2 := blockutil.AllocateBlock()
		generalutils.DownloadStream(block1,block2,resp,h.B,obj.tuple,obj.rest)
	case path=="/upload":
		// u := req.URL.RawQuery
		if req.ParseForm()!=nil { goto redirect }
		fname := req.Form.Get("path")
		src,e := os.Open(fname)
		if e!=nil { tuploaded.Execute(resp,&upload{Error:fmt.Sprint(e)}) ; return }
		defer src.Close()
		block1 := blockutil.AllocateBlock()
		block2 := blockutil.AllocateBlock()
		_,rest,tuple,e := generalutils.UploadStream(block1,block2,rand.Reader,src,h.B,3,3)
		if e!=nil { tuploaded.Execute(resp,&upload{Error:fmt.Sprint(e)}) ; return }
		// if broken { tuploaded.Execute(resp,&upload{Error:"Upload is broken"}) ; return }
		_,filepart := filepath.Split(fname)
		url := link.MakeURL(tuple,rest,filepart)
		tuploaded.Execute(resp,&upload{Url:url})
	}
	return
	redirect:
	resp.Header().Add("Location","/")
	resp.WriteHeader(303)
	return
	errresp:
	resp.WriteHeader(404)
}

