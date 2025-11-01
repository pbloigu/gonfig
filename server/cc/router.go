package cc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port      int
	Addr      string
	ForceIpv4 bool
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
	callers map[string]caller
}

func NewRouter(c Config, s service.Service) Router {
	r := &r{
		config:  c,
		service: s,
		callers: make(map[string]caller, 0),
	}
	s.RegisterIpcCallback(r.callIpc)
	return r
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
		RealmConfigs: func() []*router.RealmConfig {
			appIds := r.service.Cached().ListApplicationIds()
			configs := make([]*router.RealmConfig, len(appIds))
			for i, appId := range r.service.Cached().ListApplicationIds() {
				configs[i] = &router.RealmConfig{
					URI:            wamp.URI(appId),
					AnonymousAuth:  false,
					Authenticators: []auth.Authenticator{newAuthenticator(appId, r.service.IsAllowed)},
				}
			}
			return configs
		}(),
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

		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", r.config.Addr, r.config.Port),
			Handler: wss,
		}

		l, err := net.Listen(r.selectNetwork(), srv.Addr)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("Failed to start listener.")
		}
		r.closer = srv

		err = srv.Serve(l)
		if err != nil {
			if err == http.ErrServerClosed {
				log.Info().Msg("C&C Server shut down.")
			} else {
				log.Fatal().AnErr("error", err).Msg("Failed to start C&C router.")
			}
		}

	}()
	for _, appId := range r.service.Cached().ListApplicationIds() {
		c, err := newCaller(r.nxr, appId, r.service)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("Unable to get local caller.")
		}
		r.callers[appId] = c
	}

	log.Info().Any("port", r.config.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started C&C router.")
}

func (r r) callIpc(appId string, ipc string, args []any) ([]any, map[string]any, error) {

	// XXX: currently callers are not purged if app is deleted
	// TODO: need to pay attention to this later
	if c, ok := r.callers[appId]; ok {
		ctx := context.Background()
		r, err := c.c.Call(ctx, ipc, nil, args, nil, nil)

		if err != nil {
			return nil, nil, err
		}

		if r != nil {
			return r.Arguments, r.ArgumentsKw, nil
		}

		log.Debug().Any("result", r).Msg("Received result.")
	}
	return nil, nil, nil
}

func (r r) selectNetwork() string {
	if r.config.ForceIpv4 {
		return "tcp4"
	} else {
		return "tcp"
	}
}
