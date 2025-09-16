package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/icza/gox/gox"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/measurements"
)

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
	AddMeasurement(appId string, measurement api.Measurement)
	InitMeasurement(appId string, measurement api.Measurement)
	GetConfiguration(appId string) api.Configuration
	GetMeasurement(appId string, measurementName string) api.Measurement
	ListMeasurementValues(appId string, measurementName string, sort Sort, pagination Pargination) api.MeasurementValues
	ListMeasurements(appId string) []api.Measurement
	IsAllowed(appId string, apiKey string) bool
	Login(login string, password string) bool
	AddStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	DeleteStatusChangeTrigger(appId string)
	UpdateStatusChangeTrigger(appId string, trigger api.StatusChangeTrigger) api.StatusChangeTrigger
	GetStatusChangeTrigger(appId string) *api.StatusChangeTrigger
}

type s struct {
	m measurements.Measurements
	c configurations.Configurations
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

func New(m measurements.Measurements, c configurations.Configurations) Service {

	s := &s{
		m: m,
		c: c,
	}
	return s
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
		})
	}
	return result
}

func (s *s) DeleteApplication(id string) {
	s.c.DeleteApplication(id)
	s.m.DeleteApplication(id)
}

func (s *s) AddConfiguration(appId string, configuration api.Configuration) {
	s.c.PersistConfiguration(appId, configurations.Configuration{
		Data: configuration.Data,
	})
}

func (s *s) AddMeasurement(appId string, measurement api.Measurement) {
	s.m.PeristMeasurement(appId, measurements.Measurement{
		Name:      measurement.Name,
		LastValue: measurement.LastValue,
	})
	event.Emit(events.NewMeasurementValue{
		AppId:       appId,
		Measurement: measurement,
	})
}

func (s *s) InitMeasurement(appId string, measurement api.Measurement) {
	s.m.InitMeasurement(appId, measurement.Name)
}

func (s *s) GetConfiguration(appId string) api.Configuration {
	c := s.c.GetConfiguration(appId)
	return api.Configuration{
		Data: c.Data,
		Date: c.CreatedAt,
	}
}

func (s *s) GetMeasurement(appId string, measurementName string) api.Measurement {
	m := s.m.GetMeasurement(appId, measurementName)
	return api.Measurement{
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

func (s *s) ListMeasurementValues(appId string, measurementName string, sort Sort, pagination Pargination) api.MeasurementValues {
	m := s.m.GetMeasurement(appId, measurementName)
	if m != (measurements.Measurement{}) {
		result := api.MeasurementValues{
			Measurement: api.Measurement{
				Name:          m.Name,
				LastValue:     m.LastValue,
				LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
			},
			Values: func() []api.MeasurementValue {
				mvs := make([]api.MeasurementValue, 0)
				for _, mv := range s.m.ListMeasurementValues(m.Id, sort.Sort, string(sort.Dir), pagination.Page, pagination.Size) {
					mvs = append(mvs, api.MeasurementValue{
						Data: mv.Data,
						Time: time.Unix(int64(mv.CreatedAt), 0),
					})
				}
				return mvs
			}(),
		}
		result.Total = s.m.CountMeasurementValues(m.Id)
		result.Page = pagination.Page
		result.PageSize = pagination.Size
		return result
	} else {
		return api.MeasurementValues{}
	}
}

func (s *s) ListMeasurements(appId string) []api.Measurement {
	result := make([]api.Measurement, 0)
	for _, m := range s.m.ListMeasurements(appId) {
		result = append(result, api.Measurement{
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
