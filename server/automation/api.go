package automation

import (
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/measurements"
	"github.com/rs/zerolog/log"
)

type db struct {
	c configurations.Cached
	m measurements.Measurements
}

func (db db) LastMeasurement(appId string, measurementName string) measurements.Measurement {
	return db.m.GetMeasurement(appId, measurementName)
}

type util struct {
}

type l struct {
}

func (u util) Log() l {
	return l{}
}

func (l l) Info(msg string) {
	log.Info().Msg(msg)
}
func (l l) Debug(msg string) {
	log.Debug().Msg(msg)
}

type statusCtx struct {
	ApplicationId string
	Status        string
}
