package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/scheduler"
	"github.com/pbloigu/gonfig/server/series"
	"github.com/pbloigu/gonfig/server/util"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Cached interface {
	ListApplicationIds() []string
	ListStatusChangeActions(appId string) []api.Action
}

// Service interface with all public functions in this file
type Service interface {
	// Utility
	NewPagination(size int, defaultSize int, page int) Pargination
	NewSort(sort string, defaultSort string, dir string) Sort

	// Application
	UpdateApplication(application api.Application) api.Application
	AddApplication(application api.Application) api.Application
	GetApplication(id string) api.Application
	ListApplications() []api.Application
	DeleteApplication(id string)
	// Configuration
	AddConfiguration(appId string, configuration api.Configuration)
	GetConfiguration(appId string) api.Configuration

	// Series
	AddSeriesValue(appId string, seriesName string, value api.SeriesValue)
	InitSeries(appId string, series api.Series)
	GetSeries(appId string, seriesName string) api.Series
	ListSeriesValues(appId string, seriesName string, sort Sort, pagination Pargination) api.SeriesValues
	ListSeries(appId string) []api.Series

	// Authentication
	IsAllowed(appId string, apiKey string) bool
	Login(login string, password string) bool

	// Status change triggers
	AddStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	DeleteStatusChangeTrigger(appId string)
	UpdateStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	GetStatusChangeTrigger(appId string) *api.StatusChangeTrigger

	// Cron triggers
	GetCronTrigger(id int) *api.CronTrigger
	AddCronTrigger(trigger api.CronTrigger) api.CronTrigger
	ListCronTriggers() []api.CronTrigger
	DeleteCronTrigger(id int)
	UpdateCronTrigger(trigger api.CronTrigger) api.CronTrigger
	IsValid(request api.CronValidationRequest) bool

	Joined(appId string)
	Left(appId string)
	IsOnline(appId string) bool
	Cached() Cached
}

type s struct {
	serDb  series.SeriesDb
	c      configurations.Configurations
	online sync.Map
	sched  scheduler.Scheduler
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

func (c *c) ListStatusChangeActions(appId string) []api.Action {
	apiActions := make([]api.Action, 0)
	modelActions := c.s.c.GetStatusChangeActions(appId)
	for _, a := range modelActions {
		apiActions = append(apiActions, api.Action{
			Description: a.Description,
			Script:      a.Script,
		})
	}
	return apiActions
}

func New(seriesDb series.SeriesDb, confDb configurations.Configurations, sched scheduler.Scheduler) Service {
	s := &s{
		serDb:  seriesDb,
		c:      confDb,
		online: sync.Map{},
		sched:  sched,
	}
	s.registerCronTriggers()
	return s
}

func (s *s) IsValid(request api.CronValidationRequest) bool {
	sch, err := cron.ParseStandard(request.String())
	if err == nil {
		log.Debug().Any("next", sch.Next(time.Now())).Msg("Next invocation.")
	}
	return err == nil
}

func (s *s) registerCronTriggers() {
	tr := s.ListCronTriggers()
	for _, t := range tr {
		s.sched.RegisterCronTrigger(t)
	}
	log.Info().Any("triggers", len(tr)).Msg("Cron triggers registered.")
}

func (s *s) Cached() Cached {
	return &c{
		s: s,
	}
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
	_, ok := s.online.Load(appId)
	return ok
}

func (s *s) Joined(appId string) {
	s.online.Store(appId, time.Now())
}

func (s *s) Left(appId string) {
	s.online.Delete(appId)
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
	event.Emit(events.ApplicationAdded{
		AppId: a.Id,
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
	event.Emit(events.ApplicationDeleted{
		AppId: id,
	})
}

func (s *s) AddConfiguration(appId string, configuration api.Configuration) {
	s.c.PersistConfiguration(appId, configurations.Configuration{
		Data: configuration.Data,
	})
}

func (s *s) AddSeriesValue(appId string, seriesName string, value api.SeriesValue) {
	s.serDb.PeristSeriesValue(appId, seriesName, series.SeriesValue{
		RecordedAt: value.Recoded,
		Data:       value.Data,
	})
	event.Emit(events.NewSeriesValue{
		AppId:      appId,
		SeriesName: seriesName,
		Value:      value,
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
		Name:              m.Name,
		LastValue:         m.LastValue,
		LastValueTime:     m.LastValueTime,
		LastValueRecorded: m.LastValueRecorded,
	}
}

func (s *s) ListCronTriggers() []api.CronTrigger {
	r := make([]api.CronTrigger, 0)
	for _, ct := range s.c.ListCronTriggers() {
		r = append(r, util.CronTriggerToApi(ct))
	}
	return r
}

func (s *s) AddCronTrigger(trigger api.CronTrigger) api.CronTrigger {
	tr := s.c.PersistCronTrigger(util.CronTriggerFromApi(trigger))
	apiTr := util.CronTriggerToApi(tr)
	s.sched.RegisterCronTrigger(apiTr)
	return apiTr
}

func (s *s) DeleteCronTrigger(id int) {
	s.c.DeleteCronTrigger(id)
	s.sched.DeregisterCronTrigger(id)
}

func (s *s) UpdateCronTrigger(trigger api.CronTrigger) api.CronTrigger {
	s.c.UpdateCronTrigger(util.CronTriggerFromApi(trigger))
	s.sched.DeregisterCronTrigger(trigger.Id)
	s.sched.RegisterCronTrigger(trigger)
	tr := s.c.GetCronTrigger(trigger.Id)
	return util.CronTriggerToApi(tr)

}

func (s *s) ListSeriesValues(appId string, seriesName string, sort Sort, pagination Pargination) api.SeriesValues {
	m := s.serDb.GetSeries(appId, seriesName)
	if m != (series.Series{}) {
		result := api.SeriesValues{
			Series: api.Series{
				Name:              m.Name,
				LastValue:         m.LastValue,
				LastValueTime:     m.LastValueTime,
				LastValueRecorded: m.LastValueRecorded,
			},
			Values: func() []api.SeriesValue {
				mvs := make([]api.SeriesValue, 0)
				for _, mv := range s.serDb.ListSeriesValues(m.Id, sort.Sort, string(sort.Dir), pagination.Page, pagination.Size) {
					mvs = append(mvs, api.SeriesValue{
						Data:    mv.Data,
						Time:    mv.CreatedAt,
						Recoded: mv.RecordedAt,
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
			Name:              m.Name,
			LastValue:         m.LastValue,
			LastValueTime:     m.LastValueTime,
			LastValueRecorded: m.LastValueRecorded,
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
	tr := s.c.PersistStatusChangeTrigger(appId, util.StatusChangeTriggerFromApi(trigger))
	return util.StatusChageTriggerToApi(tr)
}

func (s *s) DeleteStatusChangeTrigger(appId string) {
	s.c.DeleteStatusChangeTrigger(appId)
}

func (s *s) UpdateStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger {
	s.c.UpdateStatusChangeTrigger(appId, util.StatusChangeTriggerFromApi(trigger))
	tr := s.c.GetStatusChangeTrigger(appId)
	return util.StatusChageTriggerToApi(tr)
}

func (s *s) GetCronTrigger(id int) *api.CronTrigger {
	tr := s.c.GetCronTrigger(id)
	if tr.Id == 0 {
		return nil
	} else {
		apiTr := util.CronTriggerToApi(tr)
		return &apiTr
	}
}

func (s *s) GetStatusChangeTrigger(appId string) *api.StatusChangeTrigger {
	tr := s.c.GetStatusChangeTrigger(appId)
	if tr.Id == 0 {
		return nil
	} else {
		apiTr := util.StatusChageTriggerToApi(tr)
		return &apiTr
	}
}
