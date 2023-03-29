package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/W0n9/t1k-sdk-go/pkg/detection"
	"github.com/W0n9/t1k-sdk-go/pkg/gosnserver"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	sReq := "POST /form.php HTTP/1.1\r\n" +
		"Host: a.com\r\n" +
		"Content-Length: 40\r\n" +
		"Content-Type: application/json\r\n\r\n" +
		"{\"name\": \"youcai\", \"password\": \"******\"}"
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer([]byte(sReq))))
	panicIf(err)

	sRsp := "HTTP/1.1 200 OK\r\n" +
		"Content-Length: 29\r\n" +
		"Content-Type: application/json\r\n\r\n" +
		"{\"err\": \"password-incorrect\"}"
	rsp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer([]byte(sRsp))), req)
	panicIf(err)

	dc := detection.New()
	detection.MakeHttpRequestInCtx(req, dc)
	detection.MakeHttpResponseInCtx(rsp, dc)

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
