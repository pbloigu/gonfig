package cc

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/server/automation"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/rs/zerolog/log"
)

type caller struct {
	c        *client.Client
	sessions map[wamp.ID]string
	lock     *sync.RWMutex
	service  automation.Service
}

func newCaller(nxr router.Router, realm string, service automation.Service) (caller, error) {
	c, err := getClient(nxr, realm)
	if err != nil {
		return caller{}, err
	} else {
		clr := caller{
			c:        c,
			sessions: make(map[wamp.ID]string),
			lock:     &sync.RWMutex{},
			service:  service,
		}

		c.Subscribe(string(wamp.MetaEventSessionOnJoin), func(wEvent *wamp.Event) {
			args := wEvent.Arguments[0].(wamp.Dict)
			session := args["session"].(wamp.ID)
			if session == c.ID() {
				// skip local subscribe
				return
			}
			auth := args["Authorization"].(string)

			appId, _, _ := getAuthDetails(auth)
			clr.lock.Lock()
			defer clr.lock.Unlock()
			clr.sessions[session] = appId
			event.Emit(events.ApplicationOnline{AppId: appId})
		}, nil)
		c.Subscribe(string(wamp.MetaEventSessionOnLeave), func(wEvent *wamp.Event) {
			session := wEvent.Arguments[0].(wamp.ID)
			if session == c.ID() {
				return
			}
			clr.lock.Lock()
			defer clr.lock.Unlock()
			appId := clr.sessions[session]
			delete(clr.sessions, session)
			event.Emit(events.ApplicationOffline{AppId: appId})

		}, nil)
		c.Subscribe(string(wamp.MetaEventRegOnCreate), func(event *wamp.Event) {
			args := event.Arguments[1].(wamp.Dict)
			invoke := args["invoke"]
			match := args["match"]
			uri := args["uri"].(wamp.URI)

			fmt.Printf("\ninvoke: %s, match: %s, uri: %s\n", reflect.TypeOf(invoke), reflect.TypeOf(match), reflect.TypeOf(uri))
		}, nil)
		return clr, nil
	}
}

func getClient(nxr router.Router, realm string) (*client.Client, error) {
	cfg := client.Config{
		Debug:         log.Debug().Enabled(),
		Realm:         realm,
		Logger:        &log.Logger,
		Serialization: client.JSON,
	}
	client, err := client.ConnectLocal(nxr, cfg)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Failed to register local client.")
		return nil, err
	}
	log.Info().Msg("Local RPC client attached.")
	return client, nil
}
