package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/pbloigu/gonfig/server/server"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

//go:embed config.sql
var config string

//go:embed data.sql
var data string

func main() {
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
		MeasurementDb:     cstr,
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

	loadConfig(os.TempDir() + "/tmp.sqlite")
	loadData(cstr)
	fmt.Print("DATA LOADED.\n")

	s.Start()
	fmt.Printf("DEV SERVER STARTED.\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	s.Stop(time.Second * 5)
	os.Remove(os.TempDir() + "/tmp.sqlite")
}

func loadConfig(dbLoc string) {
	database.New("file:///"+dbLoc+"?_pragma=foreign_keys(1)", config, "sqlite")
}

func loadData(connStr string) {
	database.New(connStr+"?multiStatements=true&parseTime=true", data, "mysql")
}
