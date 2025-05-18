package ws

// import (
// 	"errors"
// 	"fmt"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/rs/zerolog/log"
// )

// type io struct {
// 	id        string
// 	role      Role
// 	conn      *websocket.Conn
// 	newout    chan []byte
// 	terminate chan bool
// 	receiver  func([]byte, error)
// }

// func newIo(id string, role Role, conn *websocket.Conn) io {
// 	return io{
// 		id:        id,
// 		role:      role,
// 		conn:      conn,
// 		newout:    make(chan []byte),
// 		terminate: make(chan bool),
// 	}
// }

// func (i io) start() {
// 	go i.reader()
// }

// func (i io) ticker() {

// }

// func (i io) write(msg []byte) error {
// 	select {
// 	case i.newout <- msg:
// 		{
// 			return nil
// 		}
// 	case <-i.terminate:
// 		{
// 			return errors.New("Io closed.")
// 		}
// 	}
// }

// func (i io) close() {
// 	log.Debug().Any("id", i.id).Msg("Sending termination.")
// 	i.terminate <- true
// }

// func (i io) reader() {
// 	defer func() {
// 		i.conn.Close()
// 	}()
// 	i.conn.SetReadLimit(MAX_CC_MESSAGE_SIZE)

// 	i.conn.SetPongHandler(func(string) error {
// 		// Pong received, wait for another CC_PONG_WAIT for more stuff
// 		if err := i.resetReadDeadline(); err != nil {
// 			return err
// 		}
// 		log.Debug().Any("id", i.id).Msg("Pong.")
// 		return nil
// 	})
// 	for {
// 		select {
// 		case <-i.terminate:
// 			{
// 				log.Info().Any("id", i.id).Msg("Termination singal received.")
// 				i.doTerminate()
// 				go i.receiver(nil, errors.New("Io closing. Nice talkin to you, bye."))
// 				return
// 			}
// 		default:
// 			{
// 				i.resetReadDeadline()
// 				t, b, err := i.conn.ReadMessage()
// 				if err != nil {
// 					log.Error().Any("id", i.id).AnErr("error", err).Msg("Network error while reading.")
// 					go i.receiver(b, err)
// 					i.close()
// 					return
// 				} else {
// 					if t == websocket.BinaryMessage {
// 						i.receiver(b, nil)
// 					} else if t == websocket.CloseMessage {
// 						log.Info().Any("id", i.id).Msg("Received a close message. I'm done.")
// 						go i.receiver(nil, errors.New("Channel closed."))
// 						return
// 					} else {
// 						err = errors.New(fmt.Sprintf("Unsupported message type %d received.", t))
// 						log.Error().Any("id", i.id).AnErr("error", err).Msg("Unsuppoted message type.")
// 						// no need to report, just eat it up
// 					}
// 				}
// 				log.Trace().Any("id", i.id).Msg("Received message.")
// 			}
// 		}
// 	}
// }

// func (i io) writer() {
// 	ticker := time.NewTicker(CC_PING_WAIT)
// 	defer func() {
// 		ticker.Stop()
// 		i.conn.Close()
// 	}()
// 	for {
// 		select {
// 		case <-i.terminate:
// 			{

// 			}
// 		case msg, ok := <-i.newout:
// 			{
// 				if !ok {
// 					log.Error().Any("id", i.id).Msg("Output queue closed. Bye.")
// 					i.doTerminate()
// 					return
// 				}
// 				w, err := i.conn.NextWriter(websocket.BinaryMessage)
// 				if err != nil {
// 					log.Error().Any("id", i.id).AnErr("error", err).Msg("Unalbe to obtain writer. Bye.")
// 					i.doTerminate()
// 					return
// 				}

// 				log.Debug().Any("msg", msg).Msg("Sending message.")
// 				i.resetWriteDeadline()
// 				w.Write(msg)
// 				if err := w.Close(); err != nil {
// 					break
// 				}
// 			}
// 		case <-ticker.C:
// 			{
// 				if err := i.doPing(); err != nil {
// 					i.doTerminate()
// 					return
// 				}
// 			}
// 		case <-i.terminate:
// 			{
// 				log.Info().Any("id", i.id).Msg("Termination singal received.")
// 				i.doTerminate()
// 				return
// 			}
// 		}
// 	}
// }

// func (i io) doTerminate() {
// 	i.resetWriteDeadline()
// 	if err := i.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
// 		log.Error().Any("id", i.id).AnErr("error", err).Msg("Failed to write closing message. No can do.")
// 	}
// }

// func (i io) resetReadDeadline() error {
// 	// Messasge received, wait for another CC_PONG_WAIT for more stuff
// 	if err := i.conn.SetReadDeadline(time.Now().Add(CC_PONG_WAIT)); err != nil {
// 		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set read deadline.")
// 		return err
// 	} else {
// 		return nil
// 	}
// }

// func (i io) resetWriteDeadline() error {
// 	if err := i.conn.SetWriteDeadline(time.Now().Add(CC_WRITE_WAIT)); err != nil {
// 		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set write deadline. Tough luck.")
// 		return err
// 	} else {
// 		return nil
// 	}
// }

// func (i io) doPing() error {
// 	// Wait for CC_WRITE_WAIT for pong to go through
// 	if err := i.conn.SetWriteDeadline(time.Now().Add(CC_WRITE_WAIT)); err != nil {
// 		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to set write deadline. Going to write anyway.")
// 		return err
// 	}
// 	if err := i.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 		log.Error().Any("id", i.id).AnErr("error", err).Msg("Unable to write ping message.")
// 		return err
// 	}
// 	log.Debug().Any("id", i.id).Msg("Ping.")
// 	return nil
// }
