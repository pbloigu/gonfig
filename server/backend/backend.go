package backend

import (
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port int
	Addr string
}

var srv service.Service

func Start(c Config, s service.Service) {
	srv = s

	startRestApi(c)

}

func startRestApi(c Config) {
	router := gin.Default()
	hc := huma.DefaultConfig("Gonfig API", "1.0.0")

	hc.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"apiKey": {
			Type:   "http",
			Scheme: "Bearer",
		},
	}
	humaWrapper := humagin.New(router, hc)
	humaWrapper.UseMiddleware(getApiTokenAuthMiddleware(humaWrapper))

	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/configuration"), getConfiguration)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/measurement/{name}"), getMeasurement)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/measurement/{name}"), addMeasurement)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/heartbeat"), doHeartbeat)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/heartbeat"), getHeartbeat)

	go router.Run(fmt.Sprintf("%s:%d", c.Addr, c.Port))
	log.Info().Any("port", c.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started backend REST services.")
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
