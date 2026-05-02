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
	"github.com/pbloigu/gonfig/server/scripting"
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

type controller struct {
	srv service.Service
	src scripting.Runner
}

func (c controller) executeScript(ctx context.Context, input *struct {
	Body api.ScriptExecutionRequest
}) (*struct {
	Body api.ScriptExecutionResponse
}, error) {
	err := c.src.Execute(input.Body)
	r := api.ScriptExecutionResponse{
		Ok: err == nil,
		Error: func() *string {
			if err != nil {
				e := err.Error()
				return &e
			} else {
				return nil
			}
		}(),
	}
	return &struct{ Body api.ScriptExecutionResponse }{
		Body: r,
	}, nil

}

func (c controller) litsCronTriggers(ctx context.Context, input *struct{}) (*struct{ Body []api.CronTrigger }, error) {
	return &struct{ Body []api.CronTrigger }{Body: c.srv.ListCronTriggers()}, nil
}

func (c controller) addCronTrigger(ctx context.Context, input *struct {
	Body api.CronTrigger
}) (*struct{ Body api.CronTrigger }, error) {
	trigger := c.srv.AddCronTrigger(input.Body)
	return &struct{ Body api.CronTrigger }{Body: trigger}, nil
}

func (c controller) isValid(ctx context.Context, input *struct {
	Body api.CronValidationRequest `doc:"Cron expression to validate."`
}) (*struct{}, error) {
	if !c.srv.IsValid(input.Body) {
		return nil, huma.Error400BadRequest("Cron expression is not valid.")
	} else {
		return nil, nil
	}
}

func (c controller) updateCronTrigger(ctx context.Context, input *struct {
	Id   int `path:"id" doc:"Id of the cron trigger to update."`
	Body api.CronTrigger
}) (*struct{ Body api.CronTrigger }, error) {
	trigger := c.srv.GetCronTrigger(input.Id)
	if trigger == nil {
		return nil, huma.Error404NotFound("No such trigger.")
	} else {
		return &struct{ Body api.CronTrigger }{Body: *trigger}, nil
	}
}

func (c controller) getStatusChangeTrigger(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to get the status change trigger."`
}) (*struct{ Body api.StatusChangeTrigger }, error) {
	trigger := c.srv.GetStatusChangeTrigger(input.Id)

	if trigger == nil {
		return nil, huma.Error404NotFound("No such trigger.")
	} else {
		return &struct{ Body api.StatusChangeTrigger }{Body: *trigger}, nil
	}
}

func (c controller) addStatusChangeTrigger(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application to get."`
	Body api.StatusChangeTrigger
}) (*struct{ Body api.StatusChangeTrigger }, error) {
	trigger := c.srv.AddStatusChangeTrigger(input.Id, input.Body)
	return &struct{ Body api.StatusChangeTrigger }{Body: trigger}, nil
}

func (c controller) deleteStatusChangeTrigger(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{}, error) {
	c.srv.DeleteStatusChangeTrigger(input.Id)
	return &struct{}{}, nil
}

func (c controller) deleteCronTrigger(ctx context.Context, input *struct {
	Id int `path:"id" doc:"Id of the cron trigger to delete."`
}) (*struct{}, error) {
	c.srv.DeleteCronTrigger(input.Id)
	return &struct{}{}, nil
}

func (c controller) updateStatusChangeTrigger(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application to get."`
	Body api.StatusChangeTrigger
}) (*struct{ Body api.StatusChangeTrigger }, error) {
	trigger := c.srv.UpdateStatusChangeTrigger(input.Id, input.Body)
	return &struct{ Body api.StatusChangeTrigger }{Body: trigger}, nil
}

func (c controller) getApplication(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{ Body api.Application }, error) {
	app := c.srv.GetApplication(input.Id)
	if app.Id == "" {
		return nil, huma.Error404NotFound("No such application.")
	}
	return &struct{ Body api.Application }{Body: app}, nil
}

func (c controller) isOnline(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{}, error) {
	if c.srv.IsOnline(input.Id) {
		return &struct{}{}, nil
	} else {
		return nil, huma.Error404NotFound("Application is not online.")
	}
}

func (c controller) listSeries(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to get."`
}) (*struct{ Body []api.Series }, error) {
	m := c.srv.ListSeries(input.Id)
	return &struct{ Body []api.Series }{Body: m}, nil
}

func (c controller) listSeriesValues(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application to get."`
	Name string `path:"name" doc:"Name of the series."`
	Sort string `query:"sort" enum:"created,data, recorded" required:"false" doc:"Sort by." default:"created"`
	Dir  string `query:"dir" enum:"asc,desc" required:"false" doc:"Sort direction." default:"desc"`
	Page int    `query:"page" required:"false" doc:"Page number. 1-based, please." minimum:"1" default:"1"`
	Size int    `query:"size" required:"false" maximum:"50" doc:"Page size." default:"10"`
}) (*struct{ Body api.SeriesValues }, error) {
	mvs := c.srv.ListSeriesValues(input.Id,
		input.Name,
		c.srv.NewSort(input.Sort, "recorded", input.Dir),
		c.srv.NewPagination(input.Size, 10, input.Page),
	)
	return &struct{ Body api.SeriesValues }{Body: mvs}, nil
}

func (c controller) addApplication(ctx context.Context, input *struct {
	Body api.Application
}) (*struct {
	Body api.Application
}, error) {
	app := c.srv.AddApplication(input.Body)
	return &struct{ Body api.Application }{Body: app}, nil
}

func (c controller) addSeries(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to add a new series."`
	Body api.Series
}) (*struct{}, error) {
	c.srv.InitSeries(input.Id, input.Body)
	return &struct{}{}, nil
}

func (c controller) getSeries(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to get the series."`
	Name string `path:"name" doc:"Name of the series."`
}) (*struct {
	Body api.Series
}, error) {
	m := c.srv.GetSeries(input.Id, input.Name)

	response := struct {
		Body api.Series
	}{
		Body: m,
	}
	if m.Name == "" {
		return &response, huma.Error404NotFound("No such series.")
	}
	return &response, nil
}

func (c controller) updateApplication(ctx context.Context, input *struct {
	Id   string          `path:"id" doc:"Id of the application to patch."`
	Body api.Application `doc:"The application"`
}) (*struct{ Body api.Application }, error) {

	app := c.srv.UpdateApplication(input.Body)
	return &struct{ Body api.Application }{Body: app}, nil
}

func (c controller) listApplications(ctx context.Context, input *struct{}) (*struct {
	Body []api.Application
}, error) {
	response := struct {
		Body []api.Application
	}{
		Body: c.srv.ListApplications(),
	}
	return &response, nil
}

func (c controller) deleteApplication(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application to delete."`
}) (*struct{}, error) {
	c.srv.DeleteApplication(input.Id)
	return &struct{}{}, nil
}

func (c controller) addConfiguration(ctx context.Context, input *struct {
	Id   string            `path:"id" doc:"Id of the application."`
	Body api.Configuration `doc:"Configuration data"`
}) (*struct{}, error) {

	configuration := api.Configuration{}
	configuration.Data = input.Body.Data
	c.srv.AddConfiguration(input.Id, configuration)
	return &struct{}{}, nil
}

func (c controller) logout(ctx context.Context, input *struct {
	Token string `header:"Authorization"`
}) (*struct{}, error) {
	ts.remove(getToken(input.Token))
	return &struct{}{}, nil
}

func (c controller) login(ctx context.Context, input *struct {
	Body api.LoginRequest `doc:"Login credentials"`
}) (*struct {
	Body api.LoginResponse `doc:"Bearer token."`
}, error) {
	if c.srv.Login(input.Body.Username, input.Body.Password) {
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
		return nil, huma.Error401Unauthorized("You shall not pass.")
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
