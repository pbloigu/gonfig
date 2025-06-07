package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pbloigu/gonfig/server/server"
	"github.com/rs/zerolog/log"
)

var params = server.Params{
	BackedPort:   8081,
	BackendAddr:  "localhost",
	FrontendPort: 8080,
	FrontendAddr: "localhost",
	DbLoc:        "/tmp/database.sqlite",
	CcAddr:       "localhost",
	CcPort:       9000,
}

func main() {
	start := time.Now()
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	parseParams()
	parseEnv()
	s := server.New(params)
	s.Start()
	log.Info().TimeDiff("elapsedMs", time.Now(), start).Msg("System ready.")
	waitForTermination(s, ctx, stop)
}

func parseEnv() {
	if v := getEnvVariable("GONFIG_FRONTEND_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			params.FrontendPort = i
		}
	}

	if v := getEnvVariable("GONFIG_BACKEND_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			params.BackedPort = i
		}
	}

	if v := getEnvVariable("GONFIG_CC_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			params.CcPort = i
		}
	}

	if v := getEnvVariable("GONFIG_FRONTEND_ADDR"); v != "" {
		params.FrontendAddr = v
	}

	if v := getEnvVariable("GONFIG_BACKEND_ADDR"); v != "" {
		params.BackendAddr = v
	}

	if v := getEnvVariable("GONFIG_CC_ADDR"); v != "" {
		params.CcAddr = v
	}

	if v := getEnvVariable("GONFIG_DB_LOCATION"); v != "" {
		params.DbLoc = v
	}

	if v := getEnvVariable("GONFIG_MEASUREMENT_DB"); v != "" {
		params.MeasurementDb = v
	}

}

func getEnvVariable(key string) string {
	for _, e := range os.Environ() {
		keyValue := strings.Split(e, "=")
		if len(keyValue) == 2 && keyValue[0] == key {
			return keyValue[1]
		}
	}
	return ""
}

func parseParams() {
	flag.IntVar(&params.FrontendPort, "frontendPort", 8080, "Frontend listen port. Default = 8080")
	flag.StringVar(&params.FrontendAddr, "frontendAddr", "localhost", "Listen address for the frontend.")
	flag.IntVar(&params.CcPort, "ccPort", 9000, "Command channel listen port. Default = 9000")
	flag.StringVar(&params.CcAddr, "ccAddr", "localhost", "Listen address for the command channel.")
	flag.IntVar(&params.BackedPort, "backendPort", 8081, "Backend listen port. Default = 8081")
	flag.StringVar(&params.BackendAddr, "backendAddr", "localhost", "Listen address for the backend.")
	flag.StringVar(&params.DbLoc, "dbLocation", "/tmp/database.sqlite", "Location of the database. Default = /tmp/database.sqlite")
	flag.StringVar(&params.MeasurementDb, "measurementDb", "", "Database connection string for measurements. Onyl Mysql/MariaDB are supported.")

	flag.Parse()
}

func waitForTermination(s server.Server, ctx context.Context, stop context.CancelFunc) {
	<-ctx.Done()
	stop()
	log.Info().Msg("Shutting down gracefully, press Ctrl+C again to force")
	s.Stop(5 * time.Second)
	log.Info().Msg("Have a nice day. Bye.")
}
