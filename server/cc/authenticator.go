package cc

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/rs/zerolog/log"
)

type authenticator struct {
	isAllowed func(appId string, apiKey string) bool
	realm     string
}

func newAuthenticator(realm string, isAllowed func(appId string, apiKey string) bool) auth.Authenticator {
	return authenticator{isAllowed: isAllowed, realm: realm}
}

func (a authenticator) AuthMethod() string {
	return "Custom-Basic"
}

func (a authenticator) Authenticate(sid wamp.ID, details wamp.Dict, client wamp.Peer) (*wamp.Welcome, error) {
	auth := details["Authorization"].(string)
	appId, apiKey, err := getAuthDetails(auth)

	if err != nil {
		return nil, err
	} else if a.realm != appId || !a.isAllowed(appId, apiKey) {
		return nil, errors.New("Unauthorized")
	} else {
		return &wamp.Welcome{
			Details: make(wamp.Dict),
		}, nil
	}
}

func getAuthDetails(basicAuth string) (appId, apiKey string, e error) {
	src := strings.ReplaceAll(basicAuth, "Basic ", "")
	dst, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", "", err
	} else {
		split := strings.Split(string(dst), ":")
		if len(split) != 2 {
			err = errors.New("Malformed basic auth")
			log.Error().AnErr("error", err).Any("input", string(dst)).Msg("Malformed basic auth.")
			return "", "", err
		} else {
			return split[0], split[1], nil
		}
	}
}
