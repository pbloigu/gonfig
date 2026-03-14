package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

var sessions *sync.Map = &sync.Map{}

type caller struct {
	client  *client.Client
	service service.Service
	router  r
}

func (clr *caller) onJoin(wEvent *wamp.Event) {
	args := wEvent.Arguments[0].(wamp.Dict)
	session := args["session"].(wamp.ID)
	if session == clr.client.ID() {
		// skip local subscribe
		return
	}
	auth := args["Authorization"].(string)

	appId, _, err := getAuthDetails(auth)

	if err != nil {
		log.Error().AnErr("error", err).Any("appId", appId).Msg("Client found to be not authenticated upon join. Not advertising.")
		return
	}

	sessions.Store(session, appId)
	clr.service.Joined(appId)
	event.Emit(events.ApplicationOnline{AppId: appId})
	log.Debug().Any("id", session).Msg("Session established.")
}

func (clr *caller) onLeave(wEvent *wamp.Event) {
	session := wEvent.Arguments[0].(wamp.ID)
	if session == clr.client.ID() {
		return
	}

	appId, ok := sessions.LoadAndDelete(session)
	if ok {
		clr.service.Left(appId.(string))
		event.Emit(events.ApplicationOffline{AppId: appId.(string)})
	}
}

func (clr *caller) call(ipc string, timeout time.Duration, args wamp.List) (*wamp.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return clr.client.Call(ctx, ipc, nil, args, nil, nil)
}

func newCaller(client *client.Client, service service.Service) (*caller, error) {

	clr := caller{
		client:  client,
		service: service,
	}

	if err := clr.client.Subscribe(string(wamp.MetaEventSessionOnJoin), clr.onJoin, nil); err != nil {
		return &caller{}, err
	}
	if err := clr.client.Subscribe(string(wamp.MetaEventSessionOnLeave), clr.onLeave, nil); err != nil {
		return &caller{}, err
	}
	return &clr, nil
}
