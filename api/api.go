package api

import (
	"fmt"
	"time"
)

type Configuration struct {
	Data string    `json:"data"`
	Date time.Time `json:"time" readOnly:"true" required:"false"`
}

type SeriesValues struct {
	Series   Series        `json:"series"`
	Values   []SeriesValue `json:"values"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int           `json:"total"`
}

type SeriesValue struct {
	Data *string   `json:"data" maxLength:"128" required:"false" nullable:"true"`
	Time time.Time `json:"time" readOnly:"true" required:"false" format:"date-time"`
}

type Series struct {
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
	Series        []string      `json:"series" required:"false" readOnly:"true"`
	IsOnline      bool          `json:"isOnline" required:"false" readOnly:"true"`
}

type LoginRequest struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

type LoginResponse struct {
	Token  string `json:"token" required:"true"`
	Expiry int    `json:"expiry" doc:"Token exipry in seconds."`
}

type CronValidationRequest struct {
	Minute string `json:"minute" required:"true"`
	Hour   string `json:"hour" required:"true"`
	Dom    string `json:"dayOfMonth" required:"true"`
	Month  string `json:"month" required:"true"`
	Dow    string `json:"dayOfWeek" required:"true"`
}

func (r CronValidationRequest) String() string {
	return fmt.Sprintf("%s %s %s %s %s", r.Minute, r.Hour, r.Dom, r.Minute, r.Dow)
}

type CronTrigger struct {
	Id             int      `json:"id" required:"true" readOnly:"true"`
	Description    string   `json:"description" required:"true" readOnly:"false"`
	CronExpression string   `json:"cronExpression" required:"true" readOnly:"false"`
	Actions        []Action `json:"actions" required:"false" readOnly:"false"`
}

type SeriesTrigger struct {
	Id         int      `json:"id" required:"true" readOnly:"true"`
	SeriesName string   `json:"seriesName" required:"true" readOnly:"false"`
	Actions    []Action `json:"actions" required:"false" readOnly:"false"`
}

type StatusChangeTrigger struct {
	Id      int      `json:"id" required:"true" readOnly:"true"`
	Actions []Action `json:"actions" required:"false" readOnly:"false"`
}

type Action struct {
	Id          int    `json:"id" required:"true" readOnly:"true"`
	Description string `json:"description" required:"true" readOnly:"false"`
	Script      string `json:"script" required:"true" readOnly:"false"`
}
