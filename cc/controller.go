package cc

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/websocket"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

var srv service.Service

func isAllowed(authHeader string, appId string) bool {
	if authHeader != "" && appId != "" {
		for i, p := range strings.Split(authHeader, " ") {
			if i == 1 {
				return srv.IsAllowed(appId, strings.TrimSpace(p))
			}
		}
	}
	return false
}
func getApiTokenAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		if !isAllowed(ctx.Request.Header.Get(headers.Authorization), ctx.Param("id")) {
			ctx.Request.Response.Status = "403"
			ctx.Request.Context().Done()
			return
		} else {
			ctx.Next()
		}
	}
}

func getWsHandler(e *engine) func(*gin.Context) {
	return func(c *gin.Context) {
		u := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := u.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Could not start Websocket connection.")
		}
		handleConnection(conn, e)
	}
}

func handleConnection(conn *websocket.Conn, e *engine) {
	app := &channel{engine: e, conn: conn, send: make(chan api.CC, 10)}
	app.engine.register <- app

	go app.writeLoop()
	go app.readLoop()
}
