package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/detection"
	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/gosnserver"
)

var snserver *gosnserver.Server

var snserverAddr string
var listenAddr string

func wrapHandlerFunc(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dc := detection.New()

		// detect the request
		detection.MakeHttpRequestInCtx(req, dc)
		result, err := snserver.DetectRequestInCtx(dc)
		if err != nil {
			fmt.Printf("error in detection: \n%+v\n", err)
		} else {
			if result.Blocked() {
				fmt.Fprintf(w, "blocked\n")
				return
			}
		}

		rec := httptest.NewRecorder()
		f(rec, req)
		rsp := rec.Result()

		// detect the response
		detection.MakeHttpResponseInCtx(rsp, dc)
		result, err = snserver.DetectResponseInCtx(dc)
		if err != nil {
			fmt.Printf("error in detection: \n%+v\n", err)
		} else {
			if result.Blocked() {
				fmt.Fprintf(w, "blocked\n")
				return
			}
		}

		for key, values := range rsp.Header {
			w.Header()[key] = values
		}
		_, err = io.Copy(w, rsp.Body)
		rsp.Body.Close()
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(w, "hello\n")
	if err != nil {
		fmt.Printf("error writing response: %s\n", err)
	}
}

func init() {
	flag.StringVar(&snserverAddr, "s", "169.254.0.5:8000", "address of snserver")
	flag.StringVar(&listenAddr, "l", ":8090", "listen address")
	flag.Parse()
}

func main() {
	var err error
	snserver, err = gosnserver.New(snserverAddr)
	if err != nil {
		fmt.Printf("error creating snserver: %s\n", err)
		return
	}
	http.HandleFunc("/", wrapHandlerFunc(hello))
	err = http.ListenAndServe(listenAddr, nil)
	fmt.Println("server stop: ", err.Error())
}
