package main

import (
	"cc"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pbloigu/gonfig/server/backend"
	"github.com/pbloigu/gonfig/server/frontend"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

var params = struct {
	backedPort    int
	backendAddr   string
	frontendPort  int
	frontendAddr  string
	ccPort        int
	ccAddr        string
	dbLoc         string
	measurementDb string
}{
	backedPort:   8081,
	backendAddr:  "localhost",
	frontendPort: 8080,
	frontendAddr: "localhost",
	dbLoc:        "/tmp/database.sqlite",
}

func main() {
	start := time.Now()
	parseParams()
	parseEnv()
	startApis(service.New(params.dbLoc, params.measurementDb))
	log.Info().TimeDiff("elapsedMs", time.Now(), start).Msg("System ready.")
	waitForTermination()
}

func startApis(s service.Service) {
	ch := make(chan bool)

	go func() {
		frontend.Start(frontend.Config{Port: params.frontendPort, Addr: params.frontendAddr}, s)
		ch <- true
	}()

	go func() {
		backend.Start(backend.Config{Port: params.backedPort, Addr: params.backendAddr}, s)
		ch <- true
	}()

	go func() {
		cc.Start(cc.Config{Port: params.ccPort, Addr: params.ccAddr})
		ch <- true
	}()

	for range 3 {
		<-ch
	}

	log.Info().Msg("API endpoints started.")
}

func parseEnv() {
	if v := getEnvVariable("GONFIG_FRONTEND_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			params.frontendPort = i
		}
	}

	if v := getEnvVariable("GONFIG_BACKEND_PORT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			params.backedPort = i
		}
	}

	if v := getEnvVariable("GONFIG_FRONTEND_ADDR"); v != "" {
		params.frontendAddr = v
	}

	if v := getEnvVariable("GONFIG_BACKEND_ADDR"); v != "" {
		params.backendAddr = v
	}

	if v := getEnvVariable("GONFIG_DB_LOCATION"); v != "" {
		params.dbLoc = v
	}

	if v := getEnvVariable("GONFIG_MEASUREMENT_DB"); v != "" {
		params.measurementDb = v
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
	flag.IntVar(&params.frontendPort, "frontendPort", 8080, "Frontend listen port. Default = 8080")
	flag.StringVar(&params.frontendAddr, "frontendAddr", "localhost", "Listen address for the frontend.")
	flag.IntVar(&params.ccPort, "ccPort", 9000, "Command channel listen port. Default = 9000")
	flag.StringVar(&params.ccAddr, "ccAddr", "localhost", "Listen address for the command channel.")
	flag.IntVar(&params.backedPort, "backendPort", 8081, "Backend listen port. Default = 8081")
	flag.StringVar(&params.backendAddr, "backendAddr", "localhost", "Listen address for the backend.")
	flag.StringVar(&params.dbLoc, "dbLocation", "/tmp/database.sqlite", "Location of the database. Default = /tmp/database.sqlite")
	flag.StringVar(&params.measurementDb, "measurementDb", "", "Database connection string for measurements. Onyl Mysql/MariaDB are supported.")

	flag.Parse()
}

func waitForTermination() {
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
	log.Info().Msg("Killed.")
}
