package automation

import (
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type data struct {
	s service.Service
}

type ipc struct {
	s service.Service
}

type statusCtx struct {
	ApplicationId string
	Status        string
}

type util struct {
}

type logging struct {
}

func (d data) LastSeries(appId string, seriesName string) api.Series {
	return d.s.GetSeries(appId, seriesName)
}

func (d data) ListApplications() []string {
	return d.s.Cached().ListApplicationIds()
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
	i.s.CallIpc(appId, proc)
}
