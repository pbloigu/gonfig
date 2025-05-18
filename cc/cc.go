package cc

import (
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port int
	Addr string
}

func Start(c Config) {
	router := gin.Default()
	engine := newEngine()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		AllowWebSockets:  true,
	}))

	router.Use(getApiTokenAuthMiddleware())
	router.GET("/ws/:id", getWsHandler(engine))

	go engine.run()
	go router.Run(fmt.Sprintf("%s:%d", c.Addr, c.Port))

	log.Info().Any("port", c.Port).Any("userId", os.Getuid()).Any("groupId", os.Getgid()).Msg("Started backend Command services.")
}
