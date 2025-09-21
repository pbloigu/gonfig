package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/icza/gox/gox"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/series"
	"github.com/rs/zerolog/log"
)

type Cached interface {
	ListApplicationIds() []string
	GetStatusChangeActions(appId string) []api.Action
}

// Service interface with all public functions in this file
type Service interface {
	NewPagination(size int, defaultSize int, page int) Pargination
	NewSort(sort string, defaultSort string, dir string) Sort

	UpdateApplication(application api.Application) api.Application
	AddApplication(application api.Application) api.Application
	GetApplication(id string) api.Application
	ListApplications() []api.Application
	DeleteApplication(id string)
	AddConfiguration(appId string, configuration api.Configuration)
	AddSeries(appId string, series api.Series)
	InitSeries(appId string, series api.Series)
	GetConfiguration(appId string) api.Configuration
	GetSeries(appId string, seriesName string) api.Series
	ListSeriesValues(appId string, seriesName string, sort Sort, pagination Pargination) api.SeriesValues
	ListSeries(appId string) []api.Series
	IsAllowed(appId string, apiKey string) bool
	Login(login string, password string) bool
	AddStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	DeleteStatusChangeTrigger(appId string)
	UpdateStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	GetStatusChangeTrigger(appId string) *api.StatusChangeTrigger
	Joined(appId string)
	Left(appId string)
	IsOnline(appId string) bool
	Cached() Cached
	RegisterIpcCallback(callback func(appId string, ipc string))
	CallIpc(appId string, ipc string)
}

type s struct {
	serDb       series.SeriesDb
	c           configurations.Configurations
	online      map[string]time.Time
	onlineLock  sync.RWMutex
	ipcCallback func(appId string, ipc string)
}

type c struct {
	s *s
}

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

type Sort struct {
	Sort string
	Dir  Direction
}

type Pargination struct {
	Page int
	Size int
}

func (c *c) ListApplicationIds() []string {
	return c.s.c.ListApplicationIds()
}

func (c *c) GetStatusChangeActions(appId string) []api.Action {
	apiActions := make([]api.Action, 0)
	modelActions := c.s.c.GetStatusChangeActions(appId)
	for _, a := range modelActions {
		apiActions = append(apiActions, api.Action{
			Name:   a.Name,
			Script: a.Script,
		})
	}
	return apiActions
}

func New(seriesDb string, dbLoc string) Service {
	serDb, c := startDatabases(seriesDb, dbLoc)
	s := &s{
		serDb:      serDb,
		c:          c,
		online:     make(map[string]time.Time),
		onlineLock: sync.RWMutex{},
	}
	return s
}

func (s *s) Cached() Cached {
	return &c{
		s: s,
	}
}

func (s *s) RegisterIpcCallback(callback func(appId string, ipc string)) {
	s.ipcCallback = callback
}

func (s *s) CallIpc(appId string, ipc string) {
	s.ipcCallback(appId, ipc)
}

func (s *s) NewPagination(size int, defaultSize int, page int) Pargination {
	return Pargination{
		Page: func() int {
			if page < 1 {
				return 1
			} else {
				return page
			}
		}(),
		Size: func() int {
			if size <= 0 {
				return defaultSize
			} else {
				return size
			}
		}(),
	}
}

func (s *s) NewSort(sort string, defaultSort string, dir string) Sort {
	return Sort{
		Sort: func() string {
			if sort == "" {
				return defaultSort
			} else {
				return sort
			}
		}(),
		Dir: func() Direction {
			if dir == string(DESC) {
				return DESC
			} else if dir == string(ASC) {
				return ASC
			} else {
				return DESC
			}
		}(),
	}
}

func (s *s) IsOnline(appId string) bool {
	s.onlineLock.RLock()
	defer s.onlineLock.RUnlock()
	_, ok := s.online[appId]
	return ok
}

func (s *s) Joined(appId string) {
	s.onlineLock.Lock()
	defer s.onlineLock.Unlock()
	s.online[appId] = time.Now()
	event.Emit(events.ApplicationOnline{AppId: appId})
}

func (s *s) Left(appId string) {
	s.onlineLock.Lock()
	defer s.onlineLock.Unlock()
	delete(s.online, appId)
	event.Emit(events.ApplicationOffline{AppId: appId})
}

func (s *s) UpdateApplication(application api.Application) api.Application {
	s.c.UpdateApplication(configurations.Application{
		Id:   application.Id,
		Name: application.Name,
		Configuration: configurations.Configuration{
			Data: application.Configuration.Data,
		},
	})
	return s.GetApplication(application.Id)
}

func (s *s) AddApplication(application api.Application) api.Application {
	a := s.c.PersistApplication(configurations.Application{
		Id:     uuid.NewString(),
		Name:   application.Name,
		ApiKey: uuid.NewString(),
		Configuration: configurations.Configuration{
			Data: application.Configuration.Data,
		},
	})
	return api.Application{
		Id:     a.Id,
		ApiKey: a.ApiKey,
		Name:   a.Name,
		Configuration: api.Configuration{
			Data: a.Configuration.Data,
			Date: a.Configuration.CreatedAt,
		},
	}
}

func (s *s) GetApplication(id string) api.Application {
	a := s.c.GetApplication(id)
	return api.Application{
		Id:   a.Id,
		Name: a.Name,
		Configuration: api.Configuration{
			Data: a.Configuration.Data,
			Date: a.Configuration.CreatedAt,
		},
		IsOnline: s.IsOnline(a.Id),
	}
}

func (s *s) ListApplications() []api.Application {
	result := make([]api.Application, 0)
	for _, a := range s.c.ListApplications() {
		result = append(result, api.Application{
			Id:   a.Id,
			Name: a.Name,
			Configuration: api.Configuration{
				Data: a.Configuration.Data,
				Date: a.Configuration.CreatedAt,
			},
			IsOnline: s.IsOnline(a.Id),
		})
	}
	return result
}

func (s *s) DeleteApplication(id string) {
	s.c.DeleteApplication(id)
	s.serDb.DeleteApplication(id)
}

func (s *s) AddConfiguration(appId string, configuration api.Configuration) {
	s.c.PersistConfiguration(appId, configurations.Configuration{
		Data: configuration.Data,
	})
}

func (s *s) AddSeries(appId string, ser api.Series) {
	s.serDb.PeristSeries(appId, series.Series{
		Name:      ser.Name,
		LastValue: ser.LastValue,
	})
	event.Emit(events.NewSeriesValue{
		AppId:  appId,
		Series: ser,
	})
}

func (s *s) InitSeries(appId string, ser api.Series) {
	s.serDb.InitSeries(appId, ser.Name)
}

func (s *s) GetConfiguration(appId string) api.Configuration {
	c := s.c.GetConfiguration(appId)
	return api.Configuration{
		Data: c.Data,
		Date: c.CreatedAt,
	}
}

func (s *s) GetSeries(appId string, seriesName string) api.Series {
	m := s.serDb.GetSeries(appId, seriesName)
	return api.Series{
		Name:          m.Name,
		LastValue:     m.LastValue,
		LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
	}
}

func timePtr(unixTs *int) *time.Time {
	if unixTs == nil {
		return nil
	} else {
		t := time.Unix(int64(*unixTs), 0)
		return &t
	}
}

func (s *s) ListSeriesValues(appId string, seriesName string, sort Sort, pagination Pargination) api.SeriesValues {
	m := s.serDb.GetSeries(appId, seriesName)
	if m != (series.Series{}) {
		result := api.SeriesValues{
			Series: api.Series{
				Name:          m.Name,
				LastValue:     m.LastValue,
				LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
			},
			Values: func() []api.SeriesValue {
				mvs := make([]api.SeriesValue, 0)
				for _, mv := range s.serDb.ListSeriesValues(m.Id, sort.Sort, string(sort.Dir), pagination.Page, pagination.Size) {
					mvs = append(mvs, api.SeriesValue{
						Data: mv.Data,
						Time: time.Unix(int64(mv.CreatedAt), 0),
					})
				}
				return mvs
			}(),
		}
		result.Total = s.serDb.CountSeriesValues(m.Id)
		result.Page = pagination.Page
		result.PageSize = pagination.Size
		return result
	} else {
		return api.SeriesValues{}
	}
}

func (s *s) ListSeries(appId string) []api.Series {
	result := make([]api.Series, 0)
	for _, m := range s.serDb.ListSeries(appId) {
		result = append(result, api.Series{
			Name:          m.Name,
			LastValue:     m.LastValue,
			LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
		})
	}
	return result
}

func (s *s) IsAllowed(appId string, apiKey string) bool {
	return s.c.IsAllowed(appId, apiKey)
}

func (s *s) Login(login string, password string) bool {
	return s.c.Login(login, password)
}

func (s *s) AddStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger {
	s.c.PersistStatusChangeTrigger(appId, configurations.StatusChangeTrigger{
		Actions: func() []configurations.Action {
			actions := make([]configurations.Action, 0)
			for _, a := range trigger.Actions {
				actions = append(actions, configurations.Action{
					Name:   a.Name,
					Script: a.Script,
				})
			}
			return actions
		}(),
	})
	return api.StatusChangeTrigger{
		Actions: trigger.Actions,
	}
}

func (s *s) DeleteStatusChangeTrigger(appId string) {
	s.c.DeleteStatusChangeTrigger(appId)
}

func (s *s) UpdateStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger {
	s.c.UpdateStatusChangeTrigger(appId, configurations.StatusChangeTrigger{
		Actions: func() []configurations.Action {
			actions := make([]configurations.Action, 0)
			for _, a := range trigger.Actions {
				actions = append(actions, configurations.Action{
					Name:   a.Name,
					Script: a.Script,
				})
			}
			return actions
		}(),
	})
	tr := s.c.GetStatusChangeTrigger(appId)
	return api.StatusChangeTrigger{
		Actions: func() []api.Action {
			acts := make([]api.Action, 0)
			for _, a := range tr.Actions {
				acts = append(acts, api.Action{
					Name:   a.Name,
					Script: a.Script,
				})
			}
			return acts
		}(),
	}
}

func (s *s) GetStatusChangeTrigger(appId string) *api.StatusChangeTrigger {
	tr := s.c.GetStatusChangeTrigger(appId)
	if tr.Id == 0 {
		return nil
	} else {
		return &api.StatusChangeTrigger{
			Actions: func() []api.Action {
				acts := make([]api.Action, 0)
				for _, a := range tr.Actions {
					acts = append(acts, api.Action{
						Name:   a.Name,
						Script: a.Script,
					})
				}
				return acts
			}(),
		}
	}
}

func startDatabases(seriesDb string, dbLoc string) (m series.SeriesDb, c configurations.Configurations) {
	mch := make(chan series.SeriesDb)
	cch := make(chan configurations.Configurations)

	go func() {
		mch <- series.New(seriesDb)
	}()
	go func() {
		cch <- configurations.New(dbLoc)
	}()

	m = <-mch
	c = <-cch
	log.Info().Msg("Databases started.")
	return
}
