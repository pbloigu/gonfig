package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/icza/gox/gox"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/measurements"
)

var hb = make(map[string]time.Time)

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

func NewPagination(size int, defaultSize int, page int) Pargination {
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

func NewSort(sort string, defaultSort string, dir string) Sort {
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

func DoHeartbeat(appId string) {
	hb[appId] = time.Now()
}

func GetHartbeat(appId string) api.Heartbeat {
	return api.Heartbeat{
		Time: hb[appId],
	}
}

func UpdateApplication(application api.Application) api.Application {
	configurations.UpdateApplication(configurations.Application{
		Id:   application.Id,
		Name: application.Name,
		Configuration: configurations.Configuration{
			Data: application.Configuration.Data,
		},
	})
	return GetApplication(application.Id)
}

func AddApplication(application api.Application) api.Application {
	a := configurations.PersistApplication(configurations.Application{
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

func GetApplication(id string) api.Application {
	a := configurations.GetApplication(id)
	return api.Application{
		Id:   a.Id,
		Name: a.Name,
		Configuration: api.Configuration{
			Data: a.Configuration.Data,
			Date: a.Configuration.CreatedAt,
		},
	}
}

func ListApplications() []api.Application {
	result := make([]api.Application, 0)
	for _, a := range configurations.ListApplications() {
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

func DeleteApplication(id string) {
	configurations.DeleteApplication(id)
	measurements.DeleteApplication(id)
}

func AddConfiguration(appId string, configuration api.Configuration) {
	configurations.PersistConfiguration(appId, configurations.Configuration{
		Data: configuration.Data,
	})
}

func AddMeasurement(appId string, measurement api.Measurement) {
	measurements.PeristMeasurement(appId, measurements.Measurement{
		Name:      measurement.Name,
		LastValue: measurement.LastValue,
	})
}

func InitMeasurement(appId string, measurement api.Measurement) {
	measurements.InitMeasurement(appId, measurement.Name)
}

func GetConfiguration(appId string) api.Configuration {
	c := configurations.GetConfiguration(appId)
	return api.Configuration{
		Data: c.Data,
		Date: c.CreatedAt,
	}
}

func GetMeasurement(appId string, measurementName string) api.Measurement {
	m := measurements.GetMeasurement(appId, measurementName)
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

func ListMeasurementValues(appId string, measurementName string, sort Sort, pagination Pargination) api.MeasurementValues {
	m := measurements.GetMeasurement(appId, measurementName)
	if m != (measurements.Measurement{}) {
		result := api.MeasurementValues{
			Measurement: api.Measurement{
				Name:          m.Name,
				LastValue:     m.LastValue,
				LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
			},
			Values: func() []api.MeasurementValue {
				mvs := make([]api.MeasurementValue, 0)
				for _, mv := range measurements.ListMeasurementValues(m.Id, sort.Sort, string(sort.Dir), pagination.Page, pagination.Size) {
					mvs = append(mvs, api.MeasurementValue{
						Data: mv.Data,
						Time: time.Unix(int64(mv.CreatedAt), 0),
					})
				}
				return mvs
			}(),
		}
		result.Total = measurements.CountMeasurementValues(m.Id)
		result.Page = pagination.Page
		result.PageSize = pagination.Size
		return result
	} else {
		return api.MeasurementValues{}
	}

}

func ListMeasurements(appId string) []api.Measurement {
	result := make([]api.Measurement, 0)
	for _, m := range measurements.ListMeasurements(appId) {
		result = append(result, api.Measurement{
			Name:          m.Name,
			LastValue:     m.LastValue,
			LastValueTime: gox.If(m.LastValueTime != nil, timePtr(m.LastValueTime), nil),
		})
	}
	return result
}

func IsAllowed(appId string, apiKey string) bool {
	return configurations.IsAllowed(appId, apiKey)
}

func Login(login string, password string) bool {
	return configurations.Login(login, password)
}
