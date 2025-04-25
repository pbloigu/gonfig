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
	Data *string   `json:"data" maxLength:"128" required:"false" nullable:"true"`
	Time time.Time `json:"time" readOnly:"true" required:"false" format:"date-time"`
}

type Measurement struct {
	Name          string     `json:"name" requred:"true" maxLength:"128"`
	LastValue     *string    `json:"lastValue" required:"false" readOnly:"true" nullable:"true"`
	LastValueTime *time.Time `json:"lastValueTime" readOnly:"true" required:"false" format:"date-time" nullable:"true"`
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
