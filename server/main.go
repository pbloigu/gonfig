package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pbloigu/gonfig/server/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	TRACE string = "trace"
	DEBUG string = "debug"
	INFO  string = "info"
	WARN  string = "warn"
	ERROR string = "error"
)

var params = server.Params{}
var logLevels = map[string]zerolog.Level{
	"trace": zerolog.TraceLevel,
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
}

var logLevel = zerolog.InfoLevel

func main() {
	start := time.Now()
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	parseParams()
	parseEnv()
	setupLogging()
	s := server.New(params)
	s.Start()
	log.Info().TimeDiff("elapsedMs", time.Now(), start).Msg("System ready.")
	waitForTermination(s, ctx, stop)
}

func setupLogging() {
	log.Logger = zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()

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

	if v := getEnvVariable("GONFIG_SERIES_DB"); v != "" {
		params.SeriesDb = v
	}

	if v := getEnvVariable("GONFIG_LOG_LEVEL"); v != "" {
		if lvl, ok := logLevels[v]; ok {
			logLevel = lvl
		}
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
	flag.StringVar(&params.SeriesDb, "seriesDb", "", "Database connection string for time series. Onyl Mysql/MariaDB are supported.")
	flag.Func("logLevel", "Logging level, acceptable values: trace, debug, info, warn, error. Beware, debug might log sensitive stuff.", func(s string) error {
		if lvl, ok := logLevels[s]; ok {
			logLevel = lvl
			return nil
		}
		return errors.New("")
	})

	flag.Parse()
}

func waitForTermination(s server.Server, ctx context.Context, stop context.CancelFunc) {
	<-ctx.Done()
	stop()
	log.Info().Msg("Shutting down gracefully, press Ctrl+C again to force")
	s.Stop(5 * time.Second)
	log.Info().Msg("Have a nice day. Bye.")
}
