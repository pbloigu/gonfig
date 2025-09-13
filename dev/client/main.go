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
	var rpc = client.RpcFunctions{"hello": hello}
	client.NewFromConfig(log.Default(), client.Config{
		ServerHost: "localhost",
		CcPort:     9000,
		RestPort:   8081,
		AppId:      "app1",
		ApiKey:     "app1",
		CCEnabled:  true,
	}, rpc)
	fmt.Println("STARTED")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	<-c
	fmt.Println("SHUT DOWN")
}
