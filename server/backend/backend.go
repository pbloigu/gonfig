package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port      int
	Addr      string
	ForceIpv4 bool
}

type Backend interface {
	Start()
	Stop(time.Duration)
}

type backend struct {
	config Config
	http   *http.Server
	c      controller
}

func New(c Config, s service.Service) Backend {
	return &backend{
		config: c,
		c: controller{
			srv: s,
		},
	}
}

func (b *backend) Stop(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := b.http.Shutdown(ctx); err != nil {
		log.Fatal().AnErr("error", err).Msg("Server forced to shutdown.")
	}
	log.Info().Msg("Backend REST services shut down.")
}

func (b *backend) Start() {
	b.startRestApi()
}

func (b *backend) startRestApi() {
	router := gin.Default()
	hc := huma.DefaultConfig("Gonfig API", "1.0.0")

	hc.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"apiKey": {
			Type:   "http",
			Scheme: "Bearer",
		},
	}
	humaWrapper := humagin.New(router, hc)
	humaWrapper.UseMiddleware(b.c.getApiTokenAuthMiddleware(humaWrapper))

	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/configuration"), b.c.getConfiguration)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/series/{name}"), b.c.getSeries)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/series/{name}"), b.c.addSeries)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", b.config.Addr, b.config.Port),
		Handler: router,
	}
	b.http = srv

	go func() {
		l, err := net.Listen(b.selectNetwork(), srv.Addr)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("Failed to start listener.")
		}
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Fatal().AnErr("error", err).Msg("Failed to start backend")
		}
	}()
	log.Info().Any("port", b.config.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started backend REST services.")
}

func (b backend) selectNetwork() string {
	if b.config.ForceIpv4 {
		return "tcp4"
	} else {
		return "tcp"
	}
}

func def(method string, path string) huma.Operation {
	return huma.Operation{
		Method: method,
		Path:   path,
		Security: []map[string][]string{
			{"apiKey": {}},
		},
	}
}
