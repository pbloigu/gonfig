package ws

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/rs/zerolog/log"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// type command struct {
// 	C string `json:"c"`
// }

// type result struct {
// 	R string `json:"r"`
// }

// func TestChannelWriteAccept(t *testing.T) {
// 	var sc *Channel
// 	var cc *Channel

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		conn, err := upgrader.Upgrade(w, r, nil)
// 		if err != nil {
// 			log.Error().AnErr("error", err).Msg("Upgrade failed.")
// 			return
// 		}
// 		s := NewChannel(conn, "server1", SERVER)
// 		sc = &s
// 	}))
// 	defer server.Close()

// 	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
// 	client, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		t.Fatalf("could not connect to WebSocket server: %v", err)
// 	}
// 	defer client.Close()
// 	c := NewChannel(client, "client1", CLIENT)
// 	cc = &c

// 	sc.Start()
// 	cc.Start()

// 	waiter := make(chan bool)
// 	go sc.Accept(func(id string, payload any) {
// 		log.Info().Any("id", id).Any("payload", payload).Msg("Received message.")
// 		sc.Write(id, result{R: "OK"}, nil)
// 	})
// 	go cc.Accept(func(id string, payload any) {})
// 	cc.Write("", command{C: "DIE"}, func(a any) {
// 		log.Info().Any("payload", a).Msg("Received response.")
// 		waiter <- true
// 	})

// 	<-waiter
// 	time.Sleep(time.Second * 10)
// }
