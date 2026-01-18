package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	nexus "github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/stdlog"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/api"
	"github.com/rs/zerolog/log"
)

type Client interface {
	GetConfiguration() (api.Configuration, error)
	GetSeries(string) (api.Series, error)
	AddSeriesValue(string, api.SeriesValue) error
}

type client struct {
	config Config
	c      *nexus.Client
	logger stdlog.StdLog
	rpc    RpcFunctions
}

type Config struct {
	ServerHost string
	RestPort   int
	CcPort     int
	AppId      string
	ApiKey     string
	CCEnabled  bool
}

type RpcFunction func(c context.Context, w *wamp.Invocation) nexus.InvokeResult
type RpcFunctions map[string]RpcFunction

func (rpc RpcFunction) toWamp() func(c context.Context, w *wamp.Invocation) nexus.InvokeResult {
	return rpc
}

func (c client) GetConfiguration() (api.Configuration, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/application/%s/configuration",
		c.config.ServerHost,
		c.config.RestPort,
		c.config.AppId), nil)

	if err != nil {
		return api.Configuration{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return api.Configuration{}, err
	}
	defer r.Body.Close()

	d, _ := io.ReadAll(r.Body)
	config := api.Configuration{}
	err = json.Unmarshal(d, &config)
	if err != nil {
		return config, err
	}
	return config, err
}

func (c client) GetSeries(seriesName string) (api.Series, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/application/%s/series/%s",
		c.config.ServerHost,
		c.config.RestPort,
		c.config.AppId,
		seriesName), nil)
	if err != nil {
		return api.Series{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return api.Series{}, err
	}
	defer r.Body.Close()

	d, _ := io.ReadAll(r.Body)
	m := api.Series{}
	err = json.Unmarshal(d, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}
func (c client) AddSeriesValue(seriesName string, value api.SeriesValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/application/%s/series/%s",
		c.config.ServerHost,
		c.config.RestPort,
		c.config.AppId,
		seriesName), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func NewFromConfig(logger stdlog.StdLog, config Config, rpcFunctions map[string]RpcFunction) (Client, error) {
	c := client{
		logger: logger,
		config: config,
		rpc:    rpcFunctions,
	}
	if err := c.config.check(); err != nil {
		return nil, err
	}
	go c.startCc()
	return &c, nil
}

func New(logger stdlog.StdLog, rpcFunctions map[string]RpcFunction) (Client, error) {
	c := client{
		logger: logger,
		config: Config{},
		rpc:    rpcFunctions,
	}
	c.config.readConfiguration()
	if err := c.config.check(); err != nil {
		return nil, err
	}
	go c.startCc()
	return &c, nil
}

func (c *client) registerRpc() error {
	if c.rpc != nil {
		for n, f := range c.rpc {
			if err := c.c.Register(n, f.toWamp(), nil); err != nil {
				return err
			}
			log.Info().Msg(fmt.Sprintf("Registered RPC function %s", n))
		}
	}
	return nil
}

func (c *client) startCc() {
	if !c.config.CCEnabled {
		log.Info().Msg("C&C channel not enabled.")
	}
	for {
		c.connect()
		<-c.c.Done()
		log.Info().Msg("C&C lost connection to the server.")
	}
}

func (c Config) check() error {
	if c.ServerHost == "" || c.AppId == "" || c.ApiKey == "" || c.CcPort == 0 || c.RestPort == 0 {
		return errors.New("invalid configuration")
	} else {
		return nil
	}
}

func (c *client) connect() {
	for {
		cfg := nexus.Config{
			Realm:         c.config.AppId,
			Serialization: nexus.JSON,
			Logger:        c.logger,
			HelloDetails: wamp.Dict{"authmethods": []string{"Custom-Basic"}, "Authorization": "Basic " +
				base64.StdEncoding.EncodeToString([]byte(c.config.AppId+":"+c.config.ApiKey))},
		}
		cli, err := nexus.ConnectNet(context.Background(), fmt.Sprintf("ws://%s:%d/ws", c.config.ServerHost, c.config.CcPort), cfg)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Failed to estabilsh C&C connection.")
			time.Sleep(time.Second * 3)
			log.Info().Msg("Reconnecting...")
		} else {
			c.c = cli
			log.Info().Msg("C&C listener connected.")
			if err := c.registerRpc(); err != nil {
				log.Error().AnErr("error", err).Msg("Failed to register RPC endpoints.")
			}
			return
		}
	}
}

func (c *Config) readConfiguration() {
	for _, e := range os.Environ() {
		keyValue := strings.SplitN(e, "=", 2)
		if len(keyValue) == 2 {
			switch keyValue[0] {
			case "GONFIG_HOST":
				{
					c.ServerHost = keyValue[1]
				}
			case "GONFIG_APPID":
				{
					c.AppId = keyValue[1]
				}
			case "GONFIG_APIKEY":
				{
					c.ApiKey = keyValue[1]
				}
			case "REST_PORT":
				{
					i, err := strconv.Atoi(keyValue[1])
					if err == nil {
						c.RestPort = i
					}
				}
			case "CC_PORT":
				{
					i, err := strconv.Atoi(keyValue[1])
					if err == nil {
						c.CcPort = i
					}
				}
			}
		}
	}
}
