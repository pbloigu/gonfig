package automation

import (
	"context"

	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/risor-io/risor"
	"github.com/rs/zerolog/log"
)

type Runner interface {
}

type r struct {
	s service.Service
}

func New(s service.Service) Runner {
	r := &r{
		s: s,
	}
	r.registerEventListeners()
	return r
}

func (r r) registerEventListeners() {
	event.On(func(e events.NewMeasurementValue) {})
	event.On(func(e events.ApplicationOnline) { r.handleStatusChange(e.AppId, ONLINE) })
	event.On(func(e events.ApplicationOffline) { r.handleStatusChange(e.AppId, OFFLINE) })
}

func (r r) handleStatusChange(appId string, status Status) {
	log.Info().Any("application", appId).Any("status", status).Msg("Application status changed.")
	for _, a := range r.s.Cached().GetStatusChangeActions(appId) {
		go r.runStatusChangeScript(appId, a.Script, status)
	}
}

func (r r) runStatusChangeScript(appId string, script string, status Status) {
	ctx := context.Background()
	_, err := risor.Eval(ctx, script,
		risor.WithoutDefaultGlobals(),
		risor.WithGlobal("data", data(r)),
		risor.WithGlobal("util", util{}),
		risor.WithGlobal("context", statusCtx{
			ApplicationId: appId,
			Status:        string(status),
		}),
		risor.WithGlobal("ipc", ipc(r)),
		risor.WithGlobal("log", logging{}))

	if err != nil {
		log.Error().AnErr("error", err).Msg("Script execution failed.")
	}
}
