package cc

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pbloigu/gonfig/api"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type channel struct {
	engine *engine
	conn   *websocket.Conn
	send   chan api.CC
}

func (c *channel) readLoop() {
	defer func() {
		c.engine.unregister <- c
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
		c.engine.infunnel <- cc
	}
}

func (c *channel) writeLoop() {
	ticker := time.NewTicker(api.CC_PING_WAIT)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
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
