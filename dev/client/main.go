package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nexus "github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func hello(ctx context.Context, i *wamp.Invocation) nexus.InvokeResult {
	name := i.Arguments[0].(string)
	fmt.Printf("HELLO %s\n", name)
	fmt.Println("Hello. Check out my enemies in the reply.")
	return nexus.InvokeResult{
		Args:   []any{"Yoda", "Luke", "Obi-Wan"},
		Kwargs: map[string]any{"My enemies": "See the list"},
	}
}

func authUi() string {
	login := api.LoginRequest{
		Username: "admin",
		Password: "password",
	}
	b, err := json.Marshal(login)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "http://localhost:8080/login", bytes.NewBuffer(b))
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	bb, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	auth := api.LoginResponse{}
	json.Unmarshal(bb, &auth)
	return auth.Token
}

func addSelf() (appId, apiKey string) {
	token := authUi()
	app := api.Application{
		Name: "testapp",
	}
	b, err := json.Marshal(app)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "http://localhost:8080/application", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	bb, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bb, &app)
	return app.Id, app.ApiKey
}

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Logger()
	appId, apiKey := addSelf()

	var rpc = client.RpcFunctions{"hello": hello}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client.NewFromConfig(ctx, &log.Logger, client.Config{
		ServerHost: "localhost",
		CcPort:     9000,
		RestPort:   8081,
		AppId:      appId,
		ApiKey:     apiKey,
		CCEnabled:  true,
	}, rpc)
	fmt.Println("STARTED")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	<-c
	cancel()
	fmt.Println("SHUT DOWN")
}
