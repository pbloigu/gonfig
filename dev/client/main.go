package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	nexus "github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/client"
)

func hello(ctx context.Context, i *wamp.Invocation) nexus.InvokeResult {
	fmt.Printf("HELLO")
	return nexus.InvokeResult{}
}

func main() {
	cl, _ := client.NewFromConfig(log.Default(), client.Config{
		ServerHost: "localhost",
		CcPort:     9000,
		RestPort:   8081,
		AppId:      "app1",
		ApiKey:     "app1",
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	cl.RegisterRPC("hello", hello)
	<-c
	fmt.Sprintf("SHUT DOWN")
}
