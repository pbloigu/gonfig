package ws

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type msg struct {
	Id      string
	Error   error
	Payload any
}

type Role string

const (
	SERVER Role = "SERVER"
	CLIENT Role = "CLIENT"
)

const (
	MAX_CC_MESSAGE_SIZE = 1024
	CC_PONG_WAIT        = 3 * time.Second
	CC_PING_WAIT        = (CC_PONG_WAIT * 9) / 10
	CC_WRITE_WAIT       = 10 * time.Second
)

type io struct {
	id       string
	role     Role
	conn     *websocket.Conn
	receiver func([]byte, error)
	wm       sync.Mutex
}

func newIo(id string, role Role, conn *websocket.Conn) *io {
	return &io{
		id:       id,
		role:     role,
		conn:     conn,
		receiver: func(b []byte, err error) {},
	}
}

func (i *io) start() {
	go i.reader()
	go i.ticker()
}

func (i *io) write(msg []byte) error {
	i.wm.Lock()
	defer i.wm.Unlock()
	w, err := i.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unalbe to obtain writer. Bye.")
		return err
	}

	log.Debug().Any("msg", msg).Msg("Sending message.")
	i.resetWriteDeadline()
	w.Write(msg)
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// we expect write lock being held here
func (i *io) resetWriteDeadline() error {
	if err := i.conn.SetWriteDeadline(time.Now().Add(CC_WRITE_WAIT)); err != nil {
		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set write deadline. Tough luck.")
		return err
	} else {
		return nil
	}
}

func (i *io) reader() {
	defer func() {
		i.conn.Close()
	}()
	i.conn.SetReadLimit(MAX_CC_MESSAGE_SIZE)
	i.setPongHandler()
	for {
		i.resetReadDeadline()
		t, b, err := i.conn.ReadMessage()
		if err != nil {
			log.Error().Any("id", i.id).AnErr("error", err).Msg("Network error while reading.")
			go i.receiver(b, err)
			return
		} else {
			if t == websocket.BinaryMessage {
				i.receiver(b, nil)
			} else if t == websocket.CloseMessage {
				log.Info().Any("id", i.id).Msg("Received a close message. I'm done.")
				go i.receiver(nil, errors.New("Channel closed."))
				return
			} else {
				err = errors.New(fmt.Sprintf("Unsupported message type %d received.", t))
				log.Error().Any("id", i.id).AnErr("error", err).Msg("Unsuppoted message type.")
				// no need to report, just eat it up
			}
		}
		log.Trace().Any("id", i.id).Msg("Received message.")
	}
}

func (i *io) ticker() {
	ticker := time.NewTicker(CC_PING_WAIT)
	defer func() {
		ticker.Stop()
		i.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			{
				if err := i.doPing(); err != nil {
					return
				}
			}
		}
	}
}

func (i *io) setPongHandler() {
	i.conn.SetPongHandler(func(string) error {
		// Pong received, wait for another CC_PONG_WAIT for more stuff
		if err := i.resetReadDeadline(); err != nil {
			return err
		}
		log.Debug().Any("id", i.id).Msg("Pong.")
		return nil
	})
}

func (i *io) resetReadDeadline() error {
	// Messasge received, wait for another CC_PONG_WAIT for more stuff
	if err := i.conn.SetReadDeadline(time.Now().Add(CC_PONG_WAIT)); err != nil {
		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set read deadline.")
		return err
	} else {
		return nil
	}
}

func (i *io) doPing() error {
	i.wm.Lock()
	defer i.wm.Unlock()
	// Wait for CC_WRITE_WAIT for pong to go through
	if err := i.conn.SetWriteDeadline(time.Now().Add(CC_WRITE_WAIT)); err != nil {
		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set write deadline. Going to write anyway.")
		return err
	}
	if err := i.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to write ping message.")
		return err
	}
	log.Debug().Any("id", i.id).Msg("Ping.")
	return nil
}
