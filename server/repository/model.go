package repository

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

type MeasurementValue struct {
	CreatedAt time.Time
	Data      string
}

type Measurement struct {
	Id            int
	LastValueTime time.Time
	LastValue     string
	Name          string
}
