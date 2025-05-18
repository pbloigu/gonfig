package ws

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Channel struct {
	io     *io
	id     string
	writes map[string]func(any)
	reads  map[string]func(id string, payload any)
}

func NewChannel(conn *websocket.Conn, id string, role Role) Channel {
	io := newIo(id, role, conn)
	c := Channel{
		io:     io,
		id:     id,
		writes: make(map[string]func(any)),
		reads:  make(map[string]func(id string, payload any)),
	}
	c.io.receiver = c.receiver
	return c
}

func (c Channel) receiver(b []byte, err error) {

}

func (c Channel) Start() {
	go c.io.start()
	log.Info().Any("id", c.id).Msg("Channel open for e-business.")
}

func (c Channel) Write(payload any) {

}
