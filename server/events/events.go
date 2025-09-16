package events

import "github.com/pbloigu/gonfig/api"

type NewMeasurementValue struct {
	AppId       string
	Measurement api.Measurement
}

func (nmv NewMeasurementValue) Type() uint32 {
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
