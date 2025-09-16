package automation

import (
	"github.com/pbloigu/gonfig/server/measurements"
	"github.com/rs/zerolog/log"
)

type data struct {
	s *service
}

type ipc struct {
	s *service
}

type statusCtx struct {
	ApplicationId string
	Status        string
}

type util struct {
}

type logging struct {
}

func (d data) LastMeasurement(appId string, measurementName string) measurements.Measurement {
	return d.s.getMeasurement(appId, measurementName)
}

func (d data) ListApplications() []string {
	return d.s.ListApplicationIds()
}

func (u util) Log() logging {
	return logging{}
}

func (l logging) Info(msg string) {
	log.Info().Msg(msg)
}
func (l logging) Debug(msg string) {
	log.Debug().Msg(msg)
}

func (i ipc) Call(appId string, proc string) {

}
