package nats_server

import (
	"github.com/nats-io/stan.go"
	"log/slog"
	"os"
)

type Server struct {
	Conn *stan.Conn
}

func New(ipaddr string) *Server {
	var srv Server
	var err error
	*srv.Conn, err = stan.Connect(
		"middleware",
		"usr",
		stan.NatsURL(ipaddr),
	)

	if err != nil {
		slog.Error("error initializing middleware")
		os.Exit(1)
	}
	return &srv
}

//func (s *Server) NewPublisher() {
//	err := s.Conn.Publish()
//}
