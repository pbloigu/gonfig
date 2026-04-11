package websocket

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/gammazero/nexus/v3/wamp"
	"github.com/stretchr/testify/assert"
)

func TestAuthMethod(t *testing.T) {
	a := authenticator{}
	assert.Equal(t, "Custom-Basic", a.AuthMethod())
}

func TestGetAuthDetailsOk(t *testing.T) {
	appIdIn := "appId"
	apiKeyIn := "apiKey"
	appIdOut, apiKeyOut, err := getAuthDetails(fmt.Sprintf(
		"Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", appIdIn, apiKeyIn)))))
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, appIdIn, appIdOut)
	assert.Equal(t, apiKeyIn, apiKeyOut)
}

func TestGetAuthDetailsKeyMissing(t *testing.T) {
	appIdIn := "appId"
	_, _, err := getAuthDetails(fmt.Sprintf(
		"Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s", appIdIn)))))
	assert.NotNil(t, err)
}

func TestGetAuthDetailsAppIdMissing(t *testing.T) {
	apiKeyIn := "apiKey"
	_, _, err := getAuthDetails(fmt.Sprintf(
		"Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s", apiKeyIn)))))
	assert.NotNil(t, err)
}

func TestGetAuthDetailsMalformed(t *testing.T) {
	appIdIn := "appId"
	apiKeyIn := "apiKey"
	_, _, err := getAuthDetails(fmt.Sprintf(
		"Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", appIdIn, apiKeyIn)))) + "INVALID")
	assert.NotNil(t, err)
}

func TestAuthenticateOk(t *testing.T) {
	appId := "appId"
	apiKey := "apiKey"
	a := authenticator{
		isAllowed: func(appId, apiKey string) bool { return true },
		realm:     appId,
	}

	details := wamp.Dict{}
	details["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", appId, apiKey))))

	w, err := a.Authenticate(1, details, nil)
	assert.Nil(t, err)
	assert.NotNil(t, w)
}

func TestAuthenticateMalformed(t *testing.T) {
	appId := "appId"
	apiKey := "apiKey"
	a := authenticator{
		isAllowed: func(appId, apiKey string) bool { return true },
		realm:     appId,
	}

	details := wamp.Dict{}
	details["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", appId, apiKey)))+"MALFORMED")

	w, err := a.Authenticate(1, details, nil)
	assert.NotNil(t, err)
	assert.Nil(t, w)
}

func TestAuthenticateWrongRealm(t *testing.T) {
	appId := "appId"
	apiKey := "apiKey"
	a := authenticator{
		isAllowed: func(appId, apiKey string) bool { return true },
		realm:     "wrong",
	}

	details := wamp.Dict{}
	details["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", appId, apiKey))))

	w, err := a.Authenticate(1, details, nil)
	assert.NotNil(t, err)
	assert.Nil(t, w)
}

func TestAuthenticateNotAllowed(t *testing.T) {
	appId := "appId"
	apiKey := "apiKey"
	a := authenticator{
		isAllowed: func(appId, apiKey string) bool { return false },
		realm:     appId,
	}

	details := wamp.Dict{}
	details["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", appId, apiKey))))

	w, err := a.Authenticate(1, details, nil)
	assert.NotNil(t, err)
	assert.Nil(t, w)
}
