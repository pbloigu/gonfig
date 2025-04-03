package frontend

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/pbloigu/gonfig/ui"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port int
	Addr string
}

func Start(c Config) {

	router := gin.Default()
	hc := huma.DefaultConfig("Gonfig API", "1.0.0")
	hc.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"apiKey": {
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

	huma.Register(humaWrapper, defineOperation(http.MethodGet, "/application/{id}"), getApplication)
	huma.Register(humaWrapper, defineOperation(http.MethodGet, "/application/{id}/measurements/{name}"), listMeasurementValues)
	huma.Register(humaWrapper, defineOperation(http.MethodPost, "/application"), addApplication)
	huma.Register(humaWrapper, defineOperation(http.MethodPatch, "/application/{id}"), updateApplication)
	huma.Register(humaWrapper, defineOperation(http.MethodGet, "/applications"), listApplications)
	huma.Register(humaWrapper, defineOperation(http.MethodGet, "/application/{id}/measurements"), listMeasurements)
	huma.Register(humaWrapper, defineOperation(http.MethodDelete, "/application/{id}"), deleteApplication)
	huma.Register(humaWrapper, defineOperation(http.MethodPost, "/application/{id}/configuration"), addConfiguration)
	huma.Register(humaWrapper, defineOperation(http.MethodPost, "/login"), login)
	huma.Register(humaWrapper, defineOperation(http.MethodGet, "/logout"), logout)

	go router.Run(fmt.Sprintf("%s:%d", c.Addr, c.Port))
	log.Info().Any("port", c.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started frontend.")

}

func isStatic(path string) bool {
	return !(strings.HasPrefix(path, "/application") ||
		strings.HasPrefix(path, "/login") ||
		strings.HasPrefix(path, "/logout"))
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

func defineOperation(method string, path string) huma.Operation {

	if path == "/login" {
		return huma.Operation{
			Method: method,
			Path:   path,
		}
	} else {
		return huma.Operation{
			Method: method,
			Path:   path,
			Security: []map[string][]string{
				{"apiKey": {}},
			},
		}
	}

}
