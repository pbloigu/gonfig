package frontend

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-http-utils/headers"
	"github.com/google/uuid"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/service"
)

type tokenStorage struct {
	tokens map[string]int64
	lock   sync.RWMutex
}

var ts = tokenStorage{
	tokens: make(map[string]int64),
	lock:   sync.RWMutex{},
}

func (ts *tokenStorage) store(token string, timestamp int64) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	ts.tokens[token] = timestamp
}

func (ts *tokenStorage) remove(token string) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	delete(ts.tokens, token)
}

func (ts *tokenStorage) isValid(token string, expirySeconds int64) bool {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	return !(time.Now().Unix() > ts.tokens[token]+expirySeconds)
}

func getApplication(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{ Body api.Application }, error) {
	app := service.GetApplication(input.Id)
	return &struct{ Body api.Application }{Body: app}, nil
}

func listMeasurements(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{ Body []api.Measurement }, error) {
	m := service.ListMeasurements(input.Id)
	return &struct{ Body []api.Measurement }{Body: m}, nil
}

func listMeasurementValues(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application to get."`
	Name string `path:"name" doc:"Name of the measurement."`
	Sort string `query:"sort" enum:"created,value" required:"false" doc:"Sort by." default:"created"`
	Dir  string `query:"dir" enum:"asc,desc" required:"false" doc:"Sort direction." default:"desc"`
	Page int    `query:"page" required:"false" doc:"Page number. 1-based, please." minimum:"1" default:"1"`
	Size int    `query:"size" required:"false" maximum:"50" doc:"Page size." default:"10"`
}) (*struct{ Body api.MeasurementValues }, error) {
	mvs := service.ListMeasurementValues(input.Id,
		input.Name,
		service.NewSort(input.Sort, "created", input.Dir),
		service.NewPagination(input.Size, 10, input.Page),
	)
	return &struct{ Body api.MeasurementValues }{Body: mvs}, nil
}

func addApplication(ctx context.Context, input *struct {
	Body api.Application
}) (*struct {
	Body api.Application
}, error) {
	app := service.AddApplication(input.Body)
	return &struct{ Body api.Application }{Body: app}, nil
}

func addMeasurement(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to add a new measurement."`
	Body api.Measurement
}) (*struct{}, error) {
	service.InitMeasurement(input.Id, input.Body)
	return &struct{}{}, nil
}

func getMeasurement(c context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to get the measurement."`
	Name string `path:"name" doc:"Name of the measurement."`
}) (*struct {
	Body api.Measurement
}, error) {
	response := struct {
		Body api.Measurement
	}{
		Body: service.GetMeasurement(input.Id, input.Name),
	}
	return &response, nil
}

func updateApplication(c context.Context, input *struct {
	Id   string          `path:"id" doc:"Id of the application to patch."`
	Body api.Application `doc:"The application"`
}) (*struct{ Body api.Application }, error) {

	app := service.UpdateApplication(input.Body)
	return &struct{ Body api.Application }{Body: app}, nil
}

func listApplications(c context.Context, input *struct{}) (*struct {
	Body []api.Application
}, error) {
	response := struct {
		Body []api.Application
	}{
		Body: service.ListApplications(),
	}
	return &response, nil
}

func deleteApplication(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to delete."`
}) (*struct{}, error) {
	service.DeleteApplication(input.Id)
	return &struct{}{}, nil
}

func addConfiguration(c context.Context, input *struct {
	Id   string            `path:"id" doc:"Id of the application."`
	Body api.Configuration `doc:"Configuration data"`
}) (*struct{}, error) {

	configuration := api.Configuration{}
	configuration.Data = input.Body.Data
	service.AddConfiguration(input.Id, configuration)
	return &struct{}{}, nil
}

func logout(c context.Context, input *struct {
	Token string `header:"Authorization"`
}) (*struct{}, error) {
	ts.remove(getToken(input.Token))
	return &struct{}{}, nil
}

func login(c context.Context, input *struct {
	Body api.LoginRequest `doc:"Login credentials."`
}) (*struct {
	Body api.LoginResponse `doc:"Bearer token."`
}, error) {
	if service.Login(input.Body.Username, input.Body.Password) {
		r := api.LoginResponse{
			Token:  uuid.NewString(),
			Expiry: 3600,
		}
		ts.store(r.Token, time.Now().Unix())
		return &struct {
			Body api.LoginResponse "doc:\"Bearer token.\""
		}{
			Body: r,
		}, nil
	} else {
		return nil, nil
	}
}

func getToken(authHeader string) string {
	for i, p := range strings.Split(authHeader, " ") {
		if i == 1 {
			return strings.TrimSpace(p)
		}
	}
	return ""
}

func isAllowed(authHeader string) bool {
	return ts.isValid(getToken(authHeader), 3600)
}

func getApiTokenAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header(headers.Authorization)
		if len(ctx.Operation().Security) > 0 && !isAllowed(authHeader) {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		} else {
			next(ctx)
		}
	}
}
