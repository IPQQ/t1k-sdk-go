package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

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

	server, err := gosnserver.New("169.254.0.5:8000")
	panicIf(err)
	defer server.Close()
	result, err := server.DetectHttpRequest(req)
	panicIf(err)

	if result.Passed() {
		fmt.Println("Passed")
	}
	if result.Blocked() {
		fmt.Println("Blocked")
	}
}
