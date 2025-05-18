package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pbloigu/gonfig/api"
	"github.com/rs/zerolog/log"
)

type Client interface {
	GetConfiguration() (api.Configuration, error)
	GetMeasurement(string) (api.Measurement, error)
	AddMeasurement(api.Measurement) error
	Send(api.CC)
}

type client struct {
	host     string
	restPort int
	ccPort   int
	appId    string
	apiKey   string
	ccIn     chan api.CC
	ccOut    chan api.CC
	conn     *websocket.Conn
	sent     map[string]bool
	received map[string]bool
}

func (c client) GetConfiguration() (api.Configuration, error) {
	req, err := http.NewRequest("GET", c.host+"/application/"+c.appId+"/configuration", nil)
	if err != nil {
		return api.Configuration{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
	req, err := http.NewRequest("GET", c.host+"/application/"+c.appId+"/measurement/"+measurementName, nil)
	if err != nil {
		return api.Measurement{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
	req, err := http.NewRequest("POST", c.host+"/application/"+c.appId+"/measurement/"+measurement.Name, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func New() (Client, error) {

	c := client{
		ccIn:     make(chan api.CC),
		ccOut:    make(chan api.CC),
		sent:     make(map[string]bool),
		received: make(map[string]bool),
	}
	readConfiguration(&c)
	if err := check(c); err != nil {
		return nil, err
	}
	c.startCc()
	return c, nil
}

func (c client) startCc() error {
	if err := c.openCcChannel(); err != nil {
		return err
	}
	go c.ccReceiver()
	go c.ccWriter()
	return nil
}

func (c client) Send(msg api.CC) {
	msg.Id = uuid.NewString()
	c.sent[msg.Id] = true
	c.ccOut <- msg
}

func (c client) openCcChannel() error {
	log.Info().Msg("Starting command and control channel.")
	h := make(map[string][]string)
	h[headers.Authorization] = []string{c.apiKey}
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s:%d/ws/%s", c.host, c.ccPort, c.appId), h)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Could not connect to the Gonfig server.")
		return err
	}
	c.conn = conn
	return nil
}

func (c client) ccWriter() {
	ticker := time.NewTicker(api.CC_PING_WAIT)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.ccOut:
			c.conn.SetWriteDeadline(time.Now().Add(api.CC_WRITE_WAIT))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Error().Msg("Unable to set write deadline.")
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Unalbe to obtain writer.")
				return
			}
			b, err := json.Marshal(message)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Unable to serialize message.")
				return
			}
			w.Write(b)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(api.CC_WRITE_WAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error().AnErr("error", err).Msg("Unable to write ping message.")
				return
			}
		}
	}
}

func (c client) ccReceiver() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(api.MAX_CC_MESSAGE_SIZE)
	if err := c.conn.SetReadDeadline(time.Now().Add(api.CC_PONG_WAIT)); err != nil {
		log.Error().AnErr("error", err).Msg("Unable to set read deadline.")
		return
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(api.CC_PONG_WAIT)); err != nil {
			log.Error().AnErr("error", err).Msg("Unable to set read deadline.")
			return err
		}
		return nil
	})
	for {
		cc := api.CC{}
		err := c.conn.ReadJSON(&cc)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().AnErr("error", err).Msg("Client closed the channel.")
			} else {
				log.Error().AnErr("error", err).Msg("Unable to read message.")
			}
			break
		}
		c.ccIn <- cc
	}
}

func check(c client) error {
	if c.host == "" || c.appId == "" || c.apiKey == "" || c.ccPort == 0 || c.restPort == 0 {
		return errors.New("invalid configuration")
	} else {
		return nil
	}
}

func readConfiguration(c *client) {
	for _, e := range os.Environ() {
		keyValue := strings.SplitN(e, "=", 2)
		if len(keyValue) == 2 {
			switch keyValue[0] {
			case "GONFIG_HOST":
				{
					c.host = keyValue[1]
				}
			case "GONFIG_APPID":
				{
					c.appId = keyValue[1]
				}
			case "GONFIG_APIKEY":
				{
					c.apiKey = keyValue[1]
				}
			case "REST_PORT":
				{
					i, err := strconv.Atoi(keyValue[1])
					if err == nil {
						c.restPort = i
					}
				}
			case "CC_PORT":
				{
					i, err := strconv.Atoi(keyValue[1])
					if err == nil {
						c.ccPort = i
					}
				}
			}
		}
	}
}
