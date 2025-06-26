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
