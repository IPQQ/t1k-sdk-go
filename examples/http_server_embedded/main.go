package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/W0n9/t1k-sdk-go/pkg/gosnserver"
)

var snserver *gosnserver.Server

var snserverAddr string
var listenAddr string

func hello(w http.ResponseWriter, req *http.Request) {
	result, err := snserver.DetectHttpRequest(req)
	if err != nil {
		fmt.Printf("error in detection: \n%+v\n", err)
	} else {
		if result.Blocked() {
			fmt.Fprintf(w, "blocked\n")
			return
		}
	}
	_, err = fmt.Fprintf(w, "hello\n")
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
	http.HandleFunc("/", hello)
	err = http.ListenAndServe(listenAddr, nil)
	fmt.Println("server stop: ", err.Error())
}
