package cc

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port int
	Addr string
}

type Router interface {
	Start()
	Stop(time.Duration)
}

type r struct {
	config  Config
	service service.Service
	nxr     router.Router
	closer  io.Closer
	caller  caller
}

func NewRouter(c Config, s service.Service) Router {
	return &r{
		config:  c,
		service: s,
	}
}
func (r *r) Stop(timeout time.Duration) {

	wait := make(chan bool)

	go func() {
		r.closer.Close()
		log.Info().Msg("C&C router stopped accepting connections.")
		r.nxr.Close()
		log.Info().Msg("Existing C&C connections drained.")
		wait <- true
	}()

	select {
	case <-wait:
		{
			log.Info().Msg("C&C router shut down.")
		}
	case <-time.After(timeout):
		{
			log.Fatal().Msg("Server forced to shutdown.")
		}
	}
}

func (r *r) Start() {

	// Create router instance.
	routerConfig := &router.Config{
		Debug: log.Debug().Enabled(),
		RealmConfigs: []*router.RealmConfig{
			{
				URI:            wamp.URI("gonfig.cc"),
				AnonymousAuth:  false,
				Authenticators: []auth.Authenticator{newAuthenticator(r.service)},
			},
		},
	}
	nxr, err := router.NewRouter(routerConfig, &log.Logger)
	if err != nil {
		log.Fatal().AnErr("error", err)
	}
	r.nxr = nxr

	// Create and run server.
	go func() {
		wss := router.NewWebsocketServer(nxr)
		wss.EnableRequestCapture = true
		closer, err := wss.ListenAndServe(fmt.Sprintf("%s:%d", r.config.Addr, r.config.Port))
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("Failed to start C&C router.")
		}
		r.closer = closer
	}()
	r.caller, err = newCaller(nxr, "gonfig.cc", r.service)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Unable to get local caller.")
	}
	log.Info().Any("port", r.config.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started C&C router.")
}
