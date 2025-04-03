package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/pbloigu/gonfig/server/backend"
	"github.com/pbloigu/gonfig/server/frontend"
	"github.com/pbloigu/gonfig/server/repository"
	"github.com/rs/zerolog/log"
)

var params = struct {
	backedPort   int
	backendAddr  string
	frontendPort int
	frontendAddr string
	dbLoc        string
}{
	backedPort:   8081,
	backendAddr:  "localhost",
	frontendPort: 8080,
	frontendAddr: "localhost",
	dbLoc:        "/tmp/database.sqlite",
}

func main() {
	parseParams()
	repository.StartDatabase(params.dbLoc)
	frontend.Start(frontend.Config{Port: params.frontendPort, Addr: params.frontendAddr})
	backend.Start(backend.Config{Port: params.backedPort, Addr: params.backendAddr})
	waitForTermination()
}

func parseParams() {
	flag.IntVar(&params.frontendPort, "frontendPort", 8080, "Frontend listen port. Default = 8080")
	flag.StringVar(&params.frontendAddr, "frontendAddr", "localhost", "Listen address for the frontend.")
	flag.IntVar(&params.backedPort, "backendPort", 8081, "Backend listen port. Default = 8081")
	flag.StringVar(&params.backendAddr, "backendAddr", "localhost", "Listen address for the backend.")
	flag.StringVar(&params.dbLoc, "dbLocation", "/tmp/database.sqlite", "Location of the database. Default = /tmp/database.sqlite")

	flag.Parse()
}

func waitForTermination() {
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
	log.Info().Msg("Killed.")
}
