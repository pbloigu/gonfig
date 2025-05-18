package ws

// import (
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/rs/zerolog/log"
// )

// func TestIoWriteAccept(t *testing.T) {
// 	var sIo *io
// 	var cIo *io

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		conn, err := upgrader.Upgrade(w, r, nil)
// 		if err != nil {
// 			log.Error().AnErr("error", err).Msg("Upgrade failed.")
// 			return
// 		}

// 		sIo = &io{
// 			id:        "server1",
// 			role:      SERVER,
// 			conn:      conn,
// 			in:        make(chan msg),
// 			out:       make(chan msg),
// 			terminate: make(chan bool),
// 		}
// 	}))
// 	defer server.Close()
// 	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
// 	client, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		t.Fatalf("could not connect to WebSocket server: %v", err)
// 	}
// 	defer client.Close()
// 	cIo = &io{
// 		id:   "client1",
// 		role: CLIENT,
// 		conn: client,
// 		in:   make(chan msg),
// 		out:  make(chan msg),
// 	}

// 	go sIo.reader()
// 	go cIo.reader()

// 	go sIo.writer()
// 	go cIo.writer()

// 	// start purger
// 	go func() {
// 		for {
// 			select {
// 			case <-sIo.in:
// 			case <-cIo.out:
// 			}
// 		}
// 	}()

// 	for i := range 10 {
// 		cIo.out <- msg{
// 			Id:      fmt.Sprintf("message %d", i),
// 			Payload: command{C: "DIE"},
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}
// 	sIo.close()

// 	time.Sleep(10 * time.Second)
// }
