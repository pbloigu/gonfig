package cc

import (
	"errors"

	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/pbloigu/gonfig/server/service"
)

type authenticator struct {
	service service.Service
}

func newAuthenticator(s service.Service) auth.Authenticator {
	return authenticator{service: s}
}

func (a authenticator) AuthMethod() string {
	return "Custom-Basic"
}

func (a authenticator) Authenticate(sid wamp.ID, details wamp.Dict, client wamp.Peer) (*wamp.Welcome, error) {
	auth := details["Authorization"].(string)
	appId, apiKey, err := getAuthDetails(auth)

	if err != nil {
		return nil, err
	} else if !a.service.IsAllowed(appId, apiKey) {
		return nil, errors.New("Unauthorized")
	} else {
		return &wamp.Welcome{
			Details: make(wamp.Dict),
		}, nil
	}
}
