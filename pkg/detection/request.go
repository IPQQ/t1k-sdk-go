package detection

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/W0n9/t1k-sdk-go/pkg/misc"
)

type Request interface {
	Header() ([]byte, error)
	Body() (uint32, io.ReadCloser, error)
	Extra() ([]byte, error)
}

type HttpRequest struct {
	req *http.Request
	dc  *DetectionContext // this is optional
}

func MakeHttpRequest(req *http.Request) *HttpRequest {
	return &HttpRequest{
		req: req,
	}
}

func MakeHttpRequestInCtx(req *http.Request, dc *DetectionContext) *HttpRequest {
	ret := &HttpRequest{
		req: req,
		dc:  dc,
	}
	dc.Request = ret
	dc.ReqBeginTime = misc.Now()
	return ret
}

func (r *HttpRequest) Header() ([]byte, error) {
	var buf bytes.Buffer
	startLine := fmt.Sprintf("%s %s HTTP/1.1\r\n", r.req.Method, r.req.URL.RequestURI())
	_, err := buf.Write([]byte(startLine))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write([]byte(fmt.Sprintf("Host: %s\r\n", r.req.Host)))
	if err != nil {
		return nil, err
	}
	err = r.req.Header.Write(&buf)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write([]byte("\r\n"))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *HttpRequest) Body() (uint32, io.ReadCloser, error) {
	bodyBytes, err := ioutil.ReadAll(r.req.Body)
	if err != nil {
		return 0, nil, err
	}
	r.req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	return uint32(len(bodyBytes)), ioutil.NopCloser(bytes.NewReader(bodyBytes)), nil
}

func (r *HttpRequest) Extra() ([]byte, error) {
	if r.dc == nil {
		return PlaceholderRequestExtra(misc.GenUUID()), nil
	}
	return GenRequestExtra(r.dc), nil
}
