package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pbloigu/gonfig/api"
)

type Client interface {
	GetConfiguration() (api.Configuration, error)
	GetMeasurement(string) (api.Measurement, error)
	AddMeasurement(api.Measurement) error
}

type client struct {
	host   string
	appId  string
	apiKey string
}

func (c client) GetConfiguration() (api.Configuration, error) {
	req, err := http.NewRequest("GET", c.host+"/application/"+c.appId+"/configuration", nil)
	if err != nil {
		return api.Configuration{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return api.Configuration{}, err
	}
	defer r.Body.Close()

	d, _ := io.ReadAll(r.Body)
	config := api.Configuration{}
	err = json.Unmarshal(d, &config)
	if err != nil {
		return config, err
	}
	return config, err
}

func (c client) GetMeasurement(measurementName string) (api.Measurement, error) {
	req, err := http.NewRequest("GET", c.host+"/application/"+c.appId+"/measurement/"+measurementName, nil)
	if err != nil {
		return api.Measurement{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return api.Measurement{}, err
	}
	defer r.Body.Close()

	d, _ := io.ReadAll(r.Body)
	m := api.Measurement{}
	err = json.Unmarshal(d, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}
func (c client) AddMeasurement(measurement api.Measurement) error {
	b, err := json.Marshal(measurement)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.host+"/application/"+c.appId+"/measurement/"+measurement.Name, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func New() (Client, error) {
	host, appId, apiKey := getEssentials()
	c := client{
		host:   host,
		appId:  appId,
		apiKey: apiKey,
	}
	if err := check(c); err != nil {
		return nil, err
	}
	return c, nil
}

func check(c client) error {
	if c.host == "" || c.appId == "" || c.apiKey == "" {
		return errors.New("invalid configuration")
	} else {
		return nil
	}
}

func getEssentials() (host, appId, apiKey string) {
	for _, e := range os.Environ() {
		keyValue := strings.SplitN(e, "=", 2)
		if len(keyValue) == 2 {
			switch keyValue[0] {
			case "GONFIG_HOST":
				{
					host = keyValue[1]
				}
			case "GONFIG_APPID":
				{
					appId = keyValue[1]
				}
			case "GONFIG_APIKEY":
				{
					apiKey = keyValue[1]
				}
			}
		}
	}
	return host, appId, apiKey
}
