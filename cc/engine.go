package cc

import "github.com/pbloigu/gonfig/api"

type engine struct {
	// Registered apps.
	apps map[*channel]bool

	// Inbound messages from the clients.
	infunnel chan api.CC

	// Register requests from the clients.
	register chan *channel

	// Unregister requests from clients.
	unregister chan *channel
}

func newEngine() *engine {
	return &engine{
		register:   make(chan *channel),
		unregister: make(chan *channel),
		apps:       make(map[*channel]bool),
		infunnel:   make(chan api.CC),
	}
}

func (e *engine) run() {
	for {
		select {
		case app := <-e.register:
			e.apps[app] = true
		case app := <-e.unregister:
			if _, ok := e.apps[app]; ok {
				delete(e.apps, app)
				close(app.send)
			}
		case message := <-e.infunnel:
			for app := range e.apps {
				select {
				case app.send <- message:
				default:
					close(app.send)
					delete(e.apps, app)
				}
			}
		}
	}
}
