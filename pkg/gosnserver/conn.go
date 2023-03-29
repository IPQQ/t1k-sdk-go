package gosnserver

import (
	"net"
	"net/http"

	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/misc"
	"git.in.chaitin.net/patronus/t1k-sdk/sdk/go/pkg/detection"
)

type conn struct {
	socket  net.Conn
	server  *Server
	failing bool
}

func makeConn(socket net.Conn, server *Server) *conn {
	return &conn{
		socket:  socket,
		server:  server,
		failing: false,
	}
}

func (c *conn) onErr(err error) {
	if err != nil {
		// re-open socket to recover from possible error state
		c.socket.Close()
		sock, errConnect := c.server.socketFactory()
		if errConnect != nil {
			c.failing = true
		}
		c.socket = sock
	}
}

func (c *conn) Close() {
	c.socket.Close()
}

func (c *conn) DetectRequestInCtx(dc *detection.DetectionContext) (*detection.Result, error) {
	ret, err := DetectRequestInCtx(c.socket, dc)
	c.onErr(err)
	return ret, err
}

func (c *conn) DetectResponseInCtx(dc *detection.DetectionContext) (*detection.Result, error) {
	ret, err := DetectResponseInCtx(c.socket, dc)
	c.onErr(err)
	return ret, misc.ErrorWrap(err, "")
}

func (c *conn) Detect(dc *detection.DetectionContext) (*detection.Result, *detection.Result, error) {
	retReq, retRsp, err := Detect(c.socket, dc)
	c.onErr(err)
	return retReq, retRsp, misc.ErrorWrap(err, "")
}

func (c *conn) DetectHttpRequest(req *http.Request) (*detection.Result, error) {
	ret, err := DetectHttpRequest(c.socket, req)
	c.onErr(err)
	return ret, err
}

func (c *conn) DetectRequest(req detection.Request) (*detection.Result, error) {
	ret, err := DetectRequest(c.socket, req)
	c.onErr(err)
	return ret, err
}

func (c *conn) Heartbeat() {
	err := DoHeartbeat(c.socket)
	c.onErr(err)
}
