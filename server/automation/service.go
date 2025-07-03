package automation

import (
	"context"

	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/measurements"
	"github.com/risor-io/risor"
	"github.com/rs/zerolog/log"
)

type Service interface {
	OnStatusChange(appId string, status Status)
}

type service struct {
	c configurations.Cached
	m measurements.Measurements
}

func New(c configurations.Cached, m measurements.Measurements) Service {
	return &service{
		c: c,
	}
}

func (s *service) OnStatusChange(appId string, status Status) {
	for _, a := range s.c.GetStatusChangeActions(appId) {
		go s.runStatusChangeScript(appId, a.Script, status)
	}
}

func (s *service) runStatusChangeScript(appId string, script string, status Status) {
	ctx := context.Background()
	_, err := risor.Eval(ctx, script,
		risor.WithoutDefaultGlobals(),
		risor.WithGlobal("api", db{
			c: s.c,
			m: s.m,
		}),
		risor.WithGlobal("util", util{}),
		risor.WithGlobal("context", statusCtx{
			ApplicationId: appId,
			Status:        string(status),
		}))

	if err != nil {
		log.Error().AnErr("error", err).Msg("Script execution failed.")
	}
}
