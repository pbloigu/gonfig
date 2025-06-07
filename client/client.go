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
	GetMeasurement(string) (api.Measurement, error)
	AddMeasurement(api.Measurement) error
}

type client struct {
	config Config
	c      *nexus.Client
	logger stdlog.StdLog
}

type Config struct {
	ServerHost string
	RestPort   int
	CcPort     int
	AppId      string
	ApiKey     string
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

func (c client) GetMeasurement(measurementName string) (api.Measurement, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/application/%s/measurement/%s",
		c.config.ServerHost,
		c.config.RestPort,
		c.config.AppId,
		measurementName), nil)
	if err != nil {
		return api.Measurement{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return api.Measurement{}, err
	}
	defer r.Body.Close()

	d, _ := io.ReadAll(r.Body)
	m := api.Measurement{}
	err = json.Unmarshal(d, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}
func (c client) AddMeasurement(measurement api.Measurement) error {
	b, err := json.Marshal(measurement)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/application/%s/measurement/%s",
		c.config.ServerHost,
		c.config.RestPort,
		c.config.AppId,
		measurement.Name), bytes.NewBuffer(b))
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

func NewFromConfig(logger stdlog.StdLog, config Config) (Client, error) {
	c := client{
		logger: logger,
		config: config,
	}
	if err := c.config.check(); err != nil {
		return nil, err
	}
	err := c.connect()
	if err != nil {
		return nil, err
	} else {
		go c.startCc()
	}
	return c, nil
}

func New(logger stdlog.StdLog) (Client, error) {

	c := client{
		logger: logger,
		config: Config{},
	}
	c.config.readConfiguration()
	if err := c.config.check(); err != nil {
		return nil, err
	}
	err := c.connect()
	if err != nil {
		return nil, err
	} else {
		go c.startCc()
	}
	return c, nil
}

func (c client) startCc() {

	for {
		log.Info().Msg("C&C listener connected.")
		<-c.c.Done()
		log.Info().Msg("C&C lost connection to the server.")
		for {
			log.Info().Msg("Reconnecting...")
			if err := c.connect(); err != nil {
				time.Sleep(time.Second * 3)
			} else {
				break
			}
		}
	}
}

func (c Config) check() error {
	if c.ServerHost == "" || c.AppId == "" || c.ApiKey == "" || c.CcPort == 0 || c.RestPort == 0 {
		return errors.New("invalid configuration")
	} else {
		return nil
	}
}

func (c *client) connect() error {
	cfg := nexus.Config{
		Realm:         "gonfig.cc",
		Serialization: nexus.JSON,
		Logger:        c.logger,
		HelloDetails: wamp.Dict{"authmethods": []string{"Custom-Basic"}, "Authorization": "Basic " +
			base64.StdEncoding.EncodeToString([]byte(c.config.AppId+":"+c.config.ApiKey))},
	}
	cli, err := nexus.ConnectNet(context.Background(), fmt.Sprintf("ws://%s:%d/ws", c.config.ServerHost, c.config.CcPort), cfg)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Failed to estabilsh C&C connection.")
		return err
	} else {
		c.c = cli
		return nil
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
