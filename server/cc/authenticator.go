package cc

import (
	"errors"

	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
)

type authenticator struct {
	isAllowed func(appId string, apiKey string) bool
}

func newAuthenticator(isAllowed func(appId string, apiKey string) bool) auth.Authenticator {
	return authenticator{isAllowed: isAllowed}
}

func (a authenticator) AuthMethod() string {
	return "Custom-Basic"
}

func (a authenticator) Authenticate(sid wamp.ID, details wamp.Dict, client wamp.Peer) (*wamp.Welcome, error) {
	auth := details["Authorization"].(string)
	appId, apiKey, err := getAuthDetails(auth)

	if err != nil {
		return nil, err
	} else if !a.isAllowed(appId, apiKey) {
		return nil, errors.New("Unauthorized")
	} else {
		return &wamp.Welcome{
			Details: make(wamp.Dict),
		}, nil
	}
}
