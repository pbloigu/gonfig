package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/pbloigu/gonfig/server/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

//go:embed db/config/*.sql
var config embed.FS

//go:embed db/data/*.sql
var data embed.FS

func main() {

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Logger()
	ctx := context.Background()
	tc, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
	)
	if err != nil {
		panic(err)
	}
	cstr, err := tc.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	s := server.New(server.Params{
		SeriesDb:          cstr,
		DbLoc:             os.TempDir() + "/tmp.sqlite",
		BackedPort:        8081,
		BackendAddr:       "0.0.0.0",
		BackendForceIpv4:  true,
		FrontendPort:      8080,
		FrontendAddr:      "0.0.0.0",
		FrontendForceIpv4: true,
		CcPort:            9000,
		CcAddr:            "0.0.0.0",
		CcForceIpv4:       true,
	})

	s.Start()
	fmt.Printf("DEV SERVER STARTED.\n")

	loadConfig(os.TempDir() + "/tmp.sqlite")
	loadData(cstr)
	fmt.Print("DATA LOADED.\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	s.Stop(time.Second * 10)
	if err := os.Remove(os.TempDir() + "/tmp.sqlite"); err != nil {
		panic(err)
	}
	fmt.Println("DONE")
}

func loadConfig(dbLoc string) {
	sub, err := fs.Sub(config, "db/config")
	if err != nil {
		panic(err)
	}
	database.New("file:///"+dbLoc+"?_pragma=foreign_keys(1)", sub, "sqlite")
}

func loadData(connStr string) {
	sub, err := fs.Sub(data, "db/data")
	if err != nil {
		panic(err)
	}
	database.New(connStr+"?multiStatements=true&parseTime=true", sub, "mysql")
}
