package configurations

import "time"

type Application struct {
	Id            string
	ApiKey        string
	Name          string
	Hostname      string
	Ip            string
	Configuration Configuration
}

type Configuration struct {
	CreatedAt time.Time
	Data      string
}

type CronTrigger struct {
	Id             int
	Name           string
	CronExpression string
	Actions        []Action
}

type MeasurementTrigger struct {
	Id              int
	MeasurementName string
	Actions         []Action
}

type StatusChangeTrigger struct {
	Id      int
	Actions []Action
}

type Action struct {
	Name   string
	Script string
}
