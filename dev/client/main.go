package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pbloigu/gonfig/client"
)

func main() {
	client.NewFromConfig(log.Default(), client.Config{
		ServerHost: "localhost",
		CcPort:     9000,
		RestPort:   8081,
		AppId:      "app1",
		ApiKey:     "app1",
	})
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
