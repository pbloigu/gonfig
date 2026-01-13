package frontend

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/pbloigu/gonfig/ui"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port      int
	Addr      string
	ForceIpv4 bool
}

type Frontend interface {
	Start()
	Stop(time.Duration)
}

type frontend struct {
	config Config
	http   *http.Server
	c      controller
}

func New(c Config, service service.Service) Frontend {
	return &frontend{
		config: c,
		c: controller{
			srv: service,
		},
	}
}

func (f *frontend) Stop(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := f.http.Shutdown(ctx); err != nil {
		log.Fatal().AnErr("error", err).Msg("Server forced to shutdown.")
	}
	log.Info().Msg("Frontend REST services shut down.")
}

func (f *frontend) Start() {
	router := gin.Default()
	hc := huma.DefaultConfig("Gonfig API", "1.0.0")
	hc.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"token": {
			Type:   "http",
			Scheme: "Bearer",
		},
	}
	humaWrapper := humagin.New(router, hc)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		} else {
			c.Next()
		}
	})

	staticHandler(router)

	humaWrapper.UseMiddleware(getApiTokenAuthMiddleware(humaWrapper))

	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}", "getApplication"), f.c.getApplication)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/online", "isOnline"), f.c.isOnline)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/series/{name}/values", "listSeriesValues"), f.c.listSeriesValues)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/series/{name}", "getSeries"), f.c.getSeries)
	huma.Register(humaWrapper, def(http.MethodPost, "/application", "addApplication"), f.c.addApplication)
	huma.Register(humaWrapper, def(http.MethodPatch, "/application/{id}", "updateApplication"), f.c.updateApplication)
	huma.Register(humaWrapper, def(http.MethodGet, "/applications", "listApplications"), f.c.listApplications)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/series", "listSeries"), f.c.listSeries)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/series", "addSeries"), f.c.addSeries)
	huma.Register(humaWrapper, def(http.MethodDelete, "/application/{id}", "deleteApplication"), f.c.deleteApplication)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/configuration", "addConfiguration"), f.c.addConfiguration)
	huma.Register(humaWrapper, def(http.MethodPost, "/login", "login"), f.c.login)
	huma.Register(humaWrapper, def(http.MethodGet, "/logout", "logout"), f.c.logout)
	huma.Register(humaWrapper, def(http.MethodGet, "/application/{id}/trigger/status", "getStatusChangeTrigger"), f.c.getStatusChangeTrigger)
	huma.Register(humaWrapper, def(http.MethodPost, "/application/{id}/trigger/status", "addStatusChangeTrigger"), f.c.addStatusChangeTrigger)
	huma.Register(humaWrapper, def(http.MethodPut, "/application/{id}/trigger/status", "updateStatusChangeTrigger"), f.c.updateStatusChangeTrigger)
	huma.Register(humaWrapper, def(http.MethodDelete, "/application/{id}/trigger/status", "deleteStatusChangeTrigger"), f.c.deleteStatusChangeTrigger)
	huma.Register(humaWrapper, def(http.MethodGet, "/crons", "listCronTriggers"), f.c.litsCronTriggers)
	huma.Register(humaWrapper, def(http.MethodPost, "/cron", "addCronTrigger"), f.c.addCronTrigger)
	huma.Register(humaWrapper, def(http.MethodPut, "/cron/{id}", "updateCronTrigger"), f.c.updateCronTrigger)
	huma.Register(humaWrapper, def(http.MethodDelete, "/cron/{id}", "deleteCronTrigger"), f.c.deleteCronTrigger)
	huma.Register(humaWrapper, def(http.MethodPost, "/cron/expression", "isValid"), f.c.isValid)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", f.config.Addr, f.config.Port),
		Handler: router,
	}
	f.http = srv
	go func() {
		l, err := net.Listen(f.selectNetwork(), srv.Addr)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("Failed to start listener.")
		}
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Fatal().AnErr("error", err).Msg("Failed to start frontend")
		}
	}()
	log.Info().Any("port", f.config.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started frontend.")
}

func (f frontend) selectNetwork() string {
	if f.config.ForceIpv4 {
		return "tcp4"
	} else {
		return "tcp"
	}
}
func isStatic(path string) bool {
	return !(strings.HasPrefix(path, "/application") ||
		strings.HasPrefix(path, "/login") ||
		strings.HasPrefix(path, "/logout") ||
		strings.HasPrefix(path, "/cron"))
}

func staticHandler(engine *gin.Engine) {
	sub, _ := fs.Sub(ui.Build, "build")

	fileServer := http.FileServer(http.FS(sub))

	engine.Use(func(c *gin.Context) {
		if isStatic(c.Request.URL.Path) {
			// Check if the requested file exists
			_, err := fs.Stat(sub, strings.TrimPrefix(c.Request.URL.Path, "/"))
			if os.IsNotExist(err) {
				// If the file does not exist, serve index.html
				fmt.Println("File not found, serving index.html")
				c.Request.URL.Path = "index.html"
			} else {
				// Serve other static files
				fmt.Println("Serving other static files")
			}

			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	})
}

func def(method string, path string, id string) huma.Operation {

	if path == "/login" {
		return huma.Operation{
			Method:      method,
			Path:        path,
			OperationID: id,
		}
	} else {
		return huma.Operation{
			Method: method,
			Path:   path,
			Security: []map[string][]string{
				{"token": {}},
			},
			OperationID: id,
		}
	}

}
