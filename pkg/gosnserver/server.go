package gosnserver

import (
	"net"
	"net/http"
	"time"

	"github.com/W0n9/t1k-sdk-go/pkg/detection"
	"github.com/W0n9/t1k-sdk-go/pkg/misc"
)

const (
	DEFAULT_POOL_SIZE  = 8
	HEARTBEAT_INTERVAL = 20
)

type Server struct {
	socketFactory func() (net.Conn, error)
	poolCh        chan *conn
	poolSize      int
	count         int
	closeCh       chan struct{}
}

func (s *Server) newConn() error {
	sock, err := s.socketFactory()
	if err != nil {
		return err
	}
	s.count += 1
	s.poolCh <- makeConn(sock, s)
	return nil
}

func (s *Server) getConn() (*conn, error) {
	if s.count < s.poolSize {
		for i := 0; i < (s.poolSize - s.count); i++ {
			err := s.newConn()
			if err != nil {
				return nil, err
			}
		}
	}
	return <-s.poolCh, nil
}

func (s *Server) putConn(c *conn) {
	if c.failing {
		s.count -= 1
	}
	s.poolCh <- c
}

func (s *Server) broadcastHeartbeat() {
	for {
		select {
		case c := <-s.poolCh:
			c.Heartbeat()
			defer s.putConn(c)
		default:
			return
		}
	}
}

func (s *Server) runHeartbeatCo() {
	for {
		timer := time.NewTimer(HEARTBEAT_INTERVAL * time.Second)
		select {
		case <-s.closeCh:
			return
		case <-timer.C:
		}
		s.broadcastHeartbeat()
	}
}

func NewFromSocketFactoryWithPoolSize(socketFactory func() (net.Conn, error), poolSize int) (*Server, error) {
	ret := &Server{
		socketFactory: socketFactory,
		poolCh:        make(chan *conn, poolSize),
		poolSize:      poolSize,
		closeCh:       make(chan struct{}),
	}
	for i := 0; i < poolSize; i++ {
		err := ret.newConn()
		if err != nil {
			return nil, err
		}
	}
	go ret.runHeartbeatCo()
	return ret, nil
}

func NewFromSocketFactory(socketFactory func() (net.Conn, error)) (*Server, error) {
	return NewFromSocketFactoryWithPoolSize(socketFactory, DEFAULT_POOL_SIZE)
}

func NewWithPoolSize(addr string, poolSize int) (*Server, error) {
	return NewFromSocketFactoryWithPoolSize(func() (net.Conn, error) {
		return net.Dial("tcp", addr)
	}, poolSize)
}

func New(addr string) (*Server, error) {
	return NewWithPoolSize(addr, DEFAULT_POOL_SIZE)
}

func (s *Server) DetectRequestInCtx(dc *detection.DetectionContext) (*detection.Result, error) {
	c, err := s.getConn()
	if err != nil {
		return nil, err
	}
	defer s.putConn(c)
	return c.DetectRequestInCtx(dc)
}

func (s *Server) DetectResponseInCtx(dc *detection.DetectionContext) (*detection.Result, error) {
	c, err := s.getConn()
	if err != nil {
		return nil, misc.ErrorWrap(err, "")
	}
	defer s.putConn(c)
	return c.DetectResponseInCtx(dc)
}

func (s *Server) Detect(dc *detection.DetectionContext) (*detection.Result, *detection.Result, error) {
	c, err := s.getConn()
	if err != nil {
		return nil, nil, misc.ErrorWrap(err, "")
	}
	defer s.putConn(c)
	return c.Detect(dc)
}

func (s *Server) DetectHttpRequest(req *http.Request) (*detection.Result, error) {
	c, err := s.getConn()
	if err != nil {
		return nil, err
	}
	defer s.putConn(c)
	return c.DetectHttpRequest(req)
}

func (s *Server) DetectRequest(req detection.Request) (*detection.Result, error) {
	c, err := s.getConn()
	if err != nil {
		return nil, err
	}
	defer s.putConn(c)
	return c.DetectRequest(req)
}

// blocks until all pending detection is completed
func (s *Server) Close() {
	close(s.closeCh)
	for i := 0; i < s.count; i++ {
		c, err := s.getConn()
		if err != nil {
			return
		}
		c.Close()
	}
}
