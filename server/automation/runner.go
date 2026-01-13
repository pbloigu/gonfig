package automation

import (
	"context"

	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/server/cc"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/builtins"
	"github.com/risor-io/risor/object"
	"github.com/rs/zerolog/log"
)

type Runner interface {
}

type runner struct {
	s service.Service
	r cc.Router
}

func New(s service.Service, r cc.Router) Runner {
	runner := &runner{
		s: s,
		r: r,
	}

	runner.registerEventListeners()
	return runner
}

func (r runner) registerEventListeners() {
	event.On(func(e events.NewSeriesValue) {})
	event.On(func(e events.ApplicationOnline) { r.handleStatusChange(e.AppId, ONLINE) })
	event.On(func(e events.ApplicationOffline) { r.handleStatusChange(e.AppId, OFFLINE) })
	event.On(func(e events.Cron) { r.handleCron(e) })
}

func (r runner) handleCron(e events.Cron) {
	log.Info().Msg("Cron event received.")
	for _, a := range e.Actions {
		go r.runCronScript(a.Script)
	}
}

func (r runner) handleStatusChange(appId string, status Status) {
	log.Info().Any("application", appId).Any("status", status).Msg("Application status changed.")
	for _, a := range r.s.Cached().ListStatusChangeActions(appId) {
		go r.runStatusChangeScript(appId, a.Script, status)
	}
}

func (r runner) risorOpts() []risor.Option {
	return []risor.Option{
		risor.WithoutDefaultGlobals(),
		risor.WithGlobal("try", object.NewBuiltin("try", builtins.Try)),
		risor.WithGlobal("data", data{
			s: r.s,
		}),
		risor.WithGlobal("ipc", ipc{
			s: r.s,
			r: r.r,
		}),
		risor.WithGlobal("log", logging{}),
	}
}

func (r runner) runCronScript(script string) {
	ctx := context.Background()
	opts := r.risorOpts()
	_, err := risor.Eval(ctx, script, opts...)

	if err != nil {
		log.Error().AnErr("error", err).Msg("Script execution failed.")
	}
}

func (r runner) runStatusChangeScript(appId string, script string, status Status) {
	ctx := context.Background()
	opts := r.risorOpts()
	opts = append(opts, risor.WithGlobal("context", statusCtx{
		ApplicationId: appId,
		Status:        string(status),
	}))
	_, err := risor.Eval(ctx, script, opts...)

	if err != nil {
		log.Error().AnErr("error", err).Msg("Script execution failed.")
	}
}
