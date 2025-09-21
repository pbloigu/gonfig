package server

import (
	"time"

	"github.com/pbloigu/gonfig/server/automation"
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
	r automation.Runner
}

type Server interface {
	Start()
	Stop(time.Duration)
}

type Params struct {
	BackedPort        int
	BackendAddr       string
	BackendForceIpv4  bool
	FrontendPort      int
	FrontendAddr      string
	FrontendForceIpv4 bool
	CcPort            int
	CcAddr            string
	CcForceIpv4       bool
	DbLoc             string
	SeriesDb          string
}

func New(p Params) Server {
	return &server{
		p: p,
	}
}

func (s *server) Start() {

	s.s = service.New(s.p.SeriesDb, s.p.DbLoc)
	s.r = automation.New(s.s)
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
	c := make(chan bool)

	go func() {
		s.f = frontend.New(frontend.Config{Port: s.p.FrontendPort, Addr: s.p.FrontendAddr, ForceIpv4: s.p.FrontendForceIpv4}, s.s)
		s.f.Start()
		c <- true
	}()

	go func() {
		s.b = backend.New(backend.Config{Port: s.p.BackedPort, Addr: s.p.BackendAddr, ForceIpv4: s.p.BackendForceIpv4}, s.s)
		s.b.Start()
		c <- true
	}()

	go func() {
		s.c = cc.NewRouter(cc.Config{Port: s.p.CcPort, Addr: s.p.CcAddr, ForceIpv4: s.p.CcForceIpv4}, s.s)
		s.c.Start()
		c <- true
	}()

	for range 3 {
		<-c
	}

	log.Info().Msg("API endpoints started.")
}
