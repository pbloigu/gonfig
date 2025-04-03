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

func getConfiguration(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to get the configuration."`
}) (*struct {
	Body api.Configuration
}, error) {
	response := struct {
		Body api.Configuration
	}{
		Body: service.GetConfiguration(input.Id),
	}
	return &response, nil
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

func doHeartbeat(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to update heartbeat."`
}) (*struct{}, error) {
	service.DoHeartbeat(input.Id)
	return &struct{}{}, nil
}

func getHeartbeat(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to update heartbeat."`
}) (*struct {
	Body api.Heartbeat
}, error) {
	return &struct{ Body api.Heartbeat }{
		Body: service.GetHartbeat(input.Id),
	}, nil
}

func addMeasurement(c context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to add measurement."`
	Name string `path:"name" doc:"The name of the measurement."`
	Body api.Measurement
}) (*struct{}, error) {
	service.AddMeasurement(input.Id, input.Body)
	return &struct{}{}, nil
}

func isAllowed(authHeader string, appId string) bool {
	if authHeader != "" && appId != "" {
		for i, p := range strings.Split(authHeader, " ") {
			if i == 1 {
				return service.IsAllowed(appId, strings.TrimSpace(p))
			}
		}
	}
	return false
}

func getApiTokenAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !isAllowed(ctx.Header(headers.Authorization), ctx.Param("id")) {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		} else {
			next(ctx)
		}
	}
}
