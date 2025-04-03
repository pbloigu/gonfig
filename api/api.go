package api

import "time"

type Heartbeat struct {
	Time time.Time `json:"time"`
}

type Configuration struct {
	Data string    `json:"data"`
	Date time.Time `json:"time" readOnly:"true" required:"false"`
}

type MeasurementValues struct {
	Measurement Measurement        `json:"measurement"`
	Values      []MeasurementValue `json:"values"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
	Total       int                `json:"total"`
}

type MeasurementValue struct {
	Data string    `json:"data" maxLength:"128"`
	Date time.Time `json:"time" readOnly:"true" required:"false"`
}

type Measurement struct {
	Name      string           `json:"name" requred:"true"`
	LastValue MeasurementValue `json:"lastValue" required:"false"`
}

type Application struct {
	Id            string        `json:"id" readOnly:"true" required:"false"`
	ApiKey        string        `json:"apiKey" readOnly:"true" required:"false"`
	Name          string        `json:"name"`
	Hostname      string        `json:"hostname" readOnly:"true" required:"false"`
	Ip            string        `json:"ip" readOnly:"true" required:"false"`
	Configuration Configuration `json:"configuration" required:"false"`
	Measurements  []string      `json:"measurements" required:"false" readonly:"true"`
}

type LoginRequest struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

type LoginResponse struct {
	Token  string `json:"token" required:"true"`
	Expiry int    `json:"expiry" doc:"Token exipry in seconds."`
}
