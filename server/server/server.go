package server

import (
	"time"

	"github.com/pbloigu/gonfig/server/backend"
	"github.com/pbloigu/gonfig/server/cc"
	"github.com/pbloigu/gonfig/server/frontend"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type server struct {
	b backend.Backend
	f frontend.Frontend
	c cc.Router
	s service.Service
	p Params
}

type Server interface {
	Start()
	Stop(time.Duration)
}

type Params struct {
	BackedPort    int
	BackendAddr   string
	FrontendPort  int
	FrontendAddr  string
	CcPort        int
	CcAddr        string
	DbLoc         string
	MeasurementDb string
}

func New(p Params) Server {
	return &server{
		p: p,
	}
}

func (s *server) Start() {
	s.s = service.New(s.p.DbLoc, s.p.MeasurementDb)
	s.startApis()
}

func (s *server) Stop(timeout time.Duration) {
	c := make(chan bool)
	go func() {
		s.f.Stop(timeout)
		c <- true
	}()
	go func() {
		s.b.Stop(timeout)
		c <- true
	}()
	go func() {
		s.c.Stop(timeout)
		c <- true
	}()
	for range 3 {
		<-c
	}
}

func (s *server) startApis() {
	s.f = frontend.New(frontend.Config{Port: s.p.FrontendPort, Addr: s.p.FrontendAddr}, s.s)
	s.b = backend.New(backend.Config{Port: s.p.BackedPort, Addr: s.p.BackendAddr}, s.s)
	s.c = cc.NewRouter(cc.Config{Port: s.p.CcPort, Addr: s.p.CcAddr}, s.s)

	s.f.Start()
	s.b.Start()
	s.c.Start()

	log.Info().Msg("API endpoints started.")
}
