package main

import (
	"context"
	"database/sql"
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
		MeasurementDb: cstr,
		DbLoc:         os.TempDir() + "/tmp.sqlite",
		BackedPort:    8081,
		BackendAddr:   "localhost",
		FrontendPort:  8080,
		FrontendAddr:  "localhost",
		CcPort:        9000,
		CcAddr:        "localhost",
	})
	s.Start()
	fmt.Printf("DEV SERVER STARTED.\n")

	loadConfig(os.TempDir() + "/tmp.sqlite")
	loadData(cstr)
	fmt.Print("DATA LOADED.\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	s.Stop(time.Second * 5)
}

func loadConfig(dbLoc string) {
	database.New("file:///"+dbLoc+"?_pragma=foreign_keys(1)", config, "sqlite")
}

func loadData(connStr string) {
	ctx := context.Background()
	o, _ := sql.Open("mysql", connStr+"?multiStatements=true&parseTime=true")
	_, err := o.ExecContext(ctx, data)
	if err != nil {
		panic(err)
	}
}
