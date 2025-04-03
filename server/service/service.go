package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/repository"
)

var hb = make(map[string]time.Time)

func DoHeartbeat(appId string) {
	hb[appId] = time.Now()
}

func GetHartbeat(appId string) api.Heartbeat {
	return api.Heartbeat{
		Time: hb[appId],
	}
}

func UpdateApplication(application api.Application) api.Application {
	repository.UpdateApplication(repository.Application{
		Id:   application.Id,
		Name: application.Name,
		Configuration: repository.Configuration{
			Data: application.Configuration.Data,
		},
	})
	return GetApplication(application.Id)
}

func AddApplication(application api.Application) api.Application {
	a := repository.PersistApplication(repository.Application{
		Id:     uuid.NewString(),
		Name:   application.Name,
		ApiKey: uuid.NewString(),
		Configuration: repository.Configuration{
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
	a := repository.GetApplication(id)
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
	for _, a := range repository.ListApplications() {
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
	repository.DeleteApplication(id)
}

func AddConfiguration(appId string, configuration api.Configuration) {
	repository.PersistConfiguration(appId, repository.Configuration{
		Data: configuration.Data,
	})
}

func AddMeasurement(appId string, measurement api.Measurement) {
	repository.PeristMeasurement(appId, repository.Measurement{
		Name:      measurement.Name,
		LastValue: measurement.LastValue.Data,
	})
}

func GetConfiguration(appId string) api.Configuration {
	c := repository.GetConfiguration(appId)
	return api.Configuration{
		Data: c.Data,
		Date: c.CreatedAt,
	}
}

func GetMeasurement(appId string, measurementName string) api.Measurement {
	m := repository.GetMeasurement(appId, measurementName)
	return api.Measurement{
		Name: m.Name,
		LastValue: api.MeasurementValue{
			Data: m.LastValue,
			Date: m.LastValueTime,
		},
	}
}

func ListMeasurementValues(appId string, measurementName string) api.MeasurementValues {
	m := repository.GetMeasurement(appId, measurementName)
	if m != (repository.Measurement{}) {
		result := api.MeasurementValues{
			Measurement: api.Measurement{
				Name: m.Name,
				LastValue: api.MeasurementValue{
					Data: m.LastValue,
					Date: m.LastValueTime,
				},
			},
			Values: func() []api.MeasurementValue {
				mvs := make([]api.MeasurementValue, 0)
				for _, mv := range repository.ListMeasurementValues(m.Id) {
					mvs = append(mvs, api.MeasurementValue{
						Data: mv.Data,
						Date: mv.CreatedAt,
					})
				}
				return mvs
			}(),
		}
		result.Total = len(result.Values)
		result.Page = 1
		result.PageSize = result.Total
		return result
	} else {
		return api.MeasurementValues{}
	}

}

func ListMeasurements(appId string) []api.Measurement {
	result := make([]api.Measurement, 0)
	for _, m := range repository.ListMeasurements(appId) {
		result = append(result, api.Measurement{
			Name: m.Name,
			LastValue: api.MeasurementValue{
				Data: m.LastValue,
				Date: m.LastValueTime,
			},
		})
	}
	return result
}

func IsAllowed(appId string, apiKey string) bool {
	return repository.IsAllowed(appId, apiKey)
}

func Login(login string, password string) bool {
	return repository.Login(login, password)
}
