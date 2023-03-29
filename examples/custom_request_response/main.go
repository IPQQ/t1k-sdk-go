package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/detection"
	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/gosnserver"
)

type MyCustomRequest struct {
	dc *detection.DetectionContext
}

func MakeMyCustomRequestInCtx(dc *detection.DetectionContext) *MyCustomRequest {
	ret := &MyCustomRequest{
		dc: dc,
	}
	dc.Request = ret
	return ret
}

func (r *MyCustomRequest) Header() ([]byte, error) {
	return []byte(
		"POST /form.php HTTP/1.1\r\n" +
			"Host: a.com\r\n" +
			"Content-Length: 40\r\n" +
			"Content-Type: application/json\r\n\r\n",
	), nil
}

func (r *MyCustomRequest) Body() (uint32, io.ReadCloser, error) {
	body := "{\"name\": \"youcai\", \"password\": \"******\"}"
	return uint32(len(body)), ioutil.NopCloser(bytes.NewReader([]byte(body))), nil
}

func (r *MyCustomRequest) Extra() ([]byte, error) {
	return detection.GenRequestExtra(r.dc), nil
}

type MyCustomResponse struct {
	dc *detection.DetectionContext
}

func MakeMyCustomResponseInCtx(dc *detection.DetectionContext) *MyCustomResponse {
	ret := &MyCustomResponse{
		dc: dc,
	}
	dc.Response = ret
	return ret
}

func (r *MyCustomResponse) RequestHeader() ([]byte, error) {
	return r.dc.Request.Header()
}

func (r *MyCustomResponse) Header() ([]byte, error) {
	return []byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Length: 29\r\n" +
			"Content-Type: application/json\r\n\r\n",
	), nil
}

func (r *MyCustomResponse) Body() (uint32, io.ReadCloser, error) {
	body := "{\"err\": \"password-incorrect\"}"
	return uint32(len(body)), ioutil.NopCloser(bytes.NewReader([]byte(body))), nil
}

func (r *MyCustomResponse) Extra() ([]byte, error) {
	return detection.GenResponseExtra(r.dc), nil
}

func (r *MyCustomResponse) T1KContext() ([]byte, error) {
	return r.dc.T1KContext, nil
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dc := detection.New()
	MakeMyCustomRequestInCtx(dc)
	MakeMyCustomResponseInCtx(dc)

	server, err := gosnserver.New("169.254.0.5:8000")
	panicIf(err)
	defer server.Close()
	resultReq, resultRsp, err := server.Detect(dc)
	panicIf(err)

	fmt.Print("Request: ")
	if resultReq.Passed() {
		fmt.Println("Passed")
	}
	if resultReq.Blocked() {
		fmt.Println("Blocked")
	}

	fmt.Print("Response: ")
	if resultRsp.Passed() {
		fmt.Println("Passed")
	}
	if resultRsp.Blocked() {
		fmt.Println("Blocked")
	}
}
