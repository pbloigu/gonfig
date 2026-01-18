package events

import "github.com/pbloigu/gonfig/api"

type NewSeriesValue struct {
	AppId      string
	SeriesName string
	Value      api.SeriesValue
}

func (nsv NewSeriesValue) Type() uint32 {
	return 1
}

type ApplicationOnline struct {
	AppId string
}

func (ao ApplicationOnline) Type() uint32 {
	return 2
}

type ApplicationOffline struct {
	AppId string
}

func (ao ApplicationOffline) Type() uint32 {
	return 3
}

type Cron struct {
	Actions []api.Action
}

func (c Cron) Type() uint32 {
	return 4
}
