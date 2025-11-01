package cc

import (
	"sync"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type caller struct {
	c        *client.Client
	sessions map[wamp.ID]string
	lock     *sync.RWMutex
	service  service.Service
}

func (clr caller) onJoin(wEvent *wamp.Event) {
	args := wEvent.Arguments[0].(wamp.Dict)
	session := args["session"].(wamp.ID)
	if session == clr.c.ID() {
		// skip local subscribe
		return
	}
	auth := args["Authorization"].(string)

	appId, _, _ := getAuthDetails(auth)
	clr.lock.Lock()
	defer clr.lock.Unlock()
	clr.sessions[session] = appId
	clr.service.Joined(appId)
	log.Debug().Any("id", session).Msg("Session established.")
}

func (clr caller) onLeave(wEvent *wamp.Event) {
	session := wEvent.Arguments[0].(wamp.ID)
	if session == clr.c.ID() {
		return
	}
	clr.lock.Lock()
	defer clr.lock.Unlock()
	appId := clr.sessions[session]
	delete(clr.sessions, session)
	clr.service.Left(appId)
}

func newCaller(nxr router.Router, realm string, service service.Service) (caller, error) {
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

		c.Subscribe(string(wamp.MetaEventSessionOnJoin), clr.onJoin, nil)
		c.Subscribe(string(wamp.MetaEventSessionOnLeave), clr.onLeave, nil)
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
