package websocket

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/server/events"
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
	CallIpc(string, string, []any) ([]any, map[string]any, error)
}

type r struct {
	config  Config
	service service.Service
	nxr     router.Router
	closer  io.Closer
	callers sync.Map
}

func New(c Config, s service.Service) Router {
	r := &r{
		config:  c,
		service: s,
		callers: sync.Map{},
	}
	event.On(func(e events.ApplicationAdded) { r.appAdded(e.AppId) })
	event.On(func(e events.ApplicationDeleted) { r.appDeleted(e.AppId) })
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
		// realm per client so that clients can't see each other
		// app id servers as the realm id
		RealmConfigs: func() []*router.RealmConfig {
			appIds := r.service.Cached().ListApplicationIds()
			configs := make([]*router.RealmConfig, len(appIds))
			for i, appId := range r.service.Cached().ListApplicationIds() {
				configs[i] = r.appRealm(appId)
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
	// each client sits in their own realm, so we also need a caller
	// in each realm in order to be able to call the clients
	for _, appId := range r.service.Cached().ListApplicationIds() {
		r.addClient(appId)
	}

	log.Info().Any("port", r.config.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started C&C router.")
}

func (r *r) appAdded(appId string) {
	r.nxr.AddRealm(r.appRealm(appId))
	r.addClient(appId)
}

func (r *r) appDeleted(appId string) {
	if c, ok := r.callers.LoadAndDelete(appId); ok {
		c.(*caller).client.Close()
		r.nxr.RemoveRealm(wamp.URI(appId))
	}
}

func (r *r) addClient(appId string) {
	cl, err := getClient(r.nxr, appId)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Unable to create local client for app realm.")
	}
	c, err := newCaller(cl, r.service)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Unable to get local caller.")
	}
	r.callers.Store(appId, c)
}

func (r *r) appRealm(appId string) *router.RealmConfig {
	return &router.RealmConfig{
		URI:            wamp.URI(appId),
		AnonymousAuth:  false,
		Authenticators: []auth.Authenticator{newAuthenticator(appId, r.service.IsAllowed)},
	}
}

func (r *r) CallIpc(appId string, ipc string, args []any) ([]any, map[string]any, error) {
	if c, ok := r.callers.Load(appId); ok {
		r, err := c.(*caller).call(ipc, args)
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

func getClient(nxr router.Router, realm string) (*client.Client, error) {
	cfg := client.Config{
		Debug:         log.Debug().Enabled(),
		Realm:         realm,
		Logger:        &log.Logger,
		Serialization: client.JSON,
	}
	client, err := client.ConnectLocal(nxr, cfg)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Failed to register local client.")
		return nil, err
	}
	log.Info().Msg("Local RPC client attached.")
	return client, nil
}

func (r *r) selectNetwork() string {
	if r.config.ForceIpv4 {
		return "tcp4"
	} else {
		return "tcp"
	}
}
