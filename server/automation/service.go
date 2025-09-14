package automation

import (
	"context"
	"sync"
	"time"

	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/measurements"
	"github.com/risor-io/risor"
	"github.com/rs/zerolog/log"
)

type Service interface {
	OnStatusChange(appId string, status Status)
	IsOnline(appId string) bool
	Joined(appId string)
	Left(appId string)
	IsAllowed(appId string, apiKey string) bool
	ListApplicationIds() []string
}

type service struct {
	c          configurations.Configurations
	m          measurements.Measurements
	online     map[string]time.Time
	onlineLock sync.RWMutex
}

func New(m measurements.Measurements, c configurations.Configurations) Service {
	return &service{
		c:          c,
		m:          m,
		online:     make(map[string]time.Time),
		onlineLock: sync.RWMutex{},
	}
}

func (s *service) ListApplicationIds() []string {
	return s.c.ListApplicationIds()
}

func (s *service) Joined(appId string) {
	s.onlineLock.Lock()
	defer s.onlineLock.Unlock()
	s.online[appId] = time.Now()
	log.Info().Any("application", appId).Msg("Application joined.")
	go s.OnStatusChange(appId, ONLINE)
}

func (s *service) Left(appId string) {
	s.onlineLock.Lock()
	defer s.onlineLock.Unlock()
	delete(s.online, appId)
	log.Info().Any("application", appId).Msg("Application went away.")
	go s.OnStatusChange(appId, OFFLINE)
}

func (s *service) IsOnline(appId string) bool {
	s.onlineLock.RLock()
	defer s.onlineLock.RUnlock()
	_, ok := s.online[appId]
	return ok
}

func (s *service) OnStatusChange(appId string, status Status) {
	for _, a := range s.c.GetStatusChangeActions(appId) {
		go s.runStatusChangeScript(appId, a.Script, status)
	}
}
func (s *service) IsAllowed(appId string, apiKey string) bool {
	return s.c.IsAllowed(appId, apiKey)
}

func (s *service) runStatusChangeScript(appId string, script string, status Status) {
	ctx := context.Background()
	_, err := risor.Eval(ctx, script,
		risor.WithoutDefaultGlobals(),
		risor.WithGlobal("api", db{
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
