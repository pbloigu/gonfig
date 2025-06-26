package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-http-utils/headers"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/service"
)

type controller struct {
	srv service.Service
}

func (c controller) getConfiguration(ctx context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to get the configuration."`
}) (*struct {
	Body api.Configuration
}, error) {
	response := struct {
		Body api.Configuration
	}{
		Body: c.srv.GetConfiguration(input.Id),
	}
	return &response, nil
}

func (c controller) getMeasurement(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to get the measurement."`
	Name string `path:"name" doc:"Name of the measurement."`
}) (*struct {
	Body api.Measurement
}, error) {
	response := struct {
		Body api.Measurement
	}{
		Body: c.srv.GetMeasurement(input.Id, input.Name),
	}
	return &response, nil
}

func (c controller) addMeasurement(ctx context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to add measurement."`
	Name string `path:"name" doc:"The name of the measurement."`
	Body api.Measurement
}) (*struct{}, error) {
	c.srv.AddMeasurement(input.Id, input.Body)
	return &struct{}{}, nil
}

func (c controller) isAllowed(authHeader string, appId string) bool {
	if authHeader != "" && appId != "" {
		for i, p := range strings.Split(authHeader, " ") {
			if i == 1 {
				return c.srv.IsAllowed(appId, strings.TrimSpace(p))
			}
		}
	}
	return false
}

func (c controller) getApiTokenAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !c.isAllowed(ctx.Header(headers.Authorization), ctx.Param("id")) {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		} else {
			next(ctx)
		}
	}
}
