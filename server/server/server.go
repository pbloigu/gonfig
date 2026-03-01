package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pbloigu/gonfig/server/backend"
	"github.com/pbloigu/gonfig/server/frontend"
	"github.com/pbloigu/gonfig/server/scripting"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/pbloigu/gonfig/server/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type server struct {
	b   backend.Backend
	f   frontend.Frontend
	wsr websocket.Router
	s   service.Service
	p   Params
	rnr scripting.Runner
}

type Server interface {
	Start()
	Stop(context.Context)
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
	s.startWebSocket()
	s.rnr = scripting.New(s.s, s.wsr)
	s.startApis()

}

func (s *server) Stop(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		s.f.Stop(ctx)
	}()
	go func() {
		defer wg.Done()
		s.b.Stop(ctx)
	}()
	go func() {
		defer wg.Done()
		s.wsr.Stop(ctx)
	}()
	wg.Wait()
}

func setGinMode() {
	if log.Logger.GetLevel() > zerolog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debug().Msg(fmt.Sprintf("%-6s %-25s --> %s (%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers))
	}
	gin.DebugPrintFunc = func(format string, values ...any) {
		log.Debug().Msg(fmt.Sprintf(format, values...))
	}
}

func (s *server) startWebSocket() {
	s.wsr = websocket.New(websocket.Config{Port: s.p.CcPort, Addr: s.p.CcAddr, ForceIpv4: s.p.CcForceIpv4}, s.s)
	s.wsr.Start()
	log.Info().Msg("WebSocket endpoints started.")
}

func (s *server) startApis() {
	setGinMode()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.f = frontend.New(frontend.Config{Port: s.p.FrontendPort, Addr: s.p.FrontendAddr, ForceIpv4: s.p.FrontendForceIpv4}, s.s, s.rnr)
		s.f.Start()
	}()

	go func() {
		defer wg.Done()
		s.b = backend.New(backend.Config{Port: s.p.BackedPort, Addr: s.p.BackendAddr, ForceIpv4: s.p.BackendForceIpv4}, s.s)
		s.b.Start()
	}()

	wg.Wait()

	log.Info().Msg("API endpoints started.")
}
