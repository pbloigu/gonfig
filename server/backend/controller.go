package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-http-utils/headers"
	"github.com/pbloigu/gonfig/api"
)

func getConfiguration(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to get the configuration."`
}) (*struct {
	Body api.Configuration
}, error) {
	response := struct {
		Body api.Configuration
	}{
		Body: srv.GetConfiguration(input.Id),
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
		Body: srv.GetMeasurement(input.Id, input.Name),
	}
	return &response, nil
}

func doHeartbeat(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to update heartbeat."`
}) (*struct{}, error) {
	srv.DoHeartbeat(input.Id)
	return &struct{}{}, nil
}

func getHeartbeat(c context.Context, input *struct {
	Id string `path:"id" doc:"Id of the application for which to update heartbeat."`
}) (*struct {
	Body api.Heartbeat
}, error) {
	return &struct{ Body api.Heartbeat }{
		Body: srv.GetHartbeat(input.Id),
	}, nil
}

func addMeasurement(c context.Context, input *struct {
	Id   string `path:"id" doc:"Id of the application for which to add measurement."`
	Name string `path:"name" doc:"The name of the measurement."`
	Body api.Measurement
}) (*struct{}, error) {
	srv.AddMeasurement(input.Id, input.Body)
	return &struct{}{}, nil
}

func isAllowed(authHeader string, appId string) bool {
	if authHeader != "" && appId != "" {
		for i, p := range strings.Split(authHeader, " ") {
			if i == 1 {
				return srv.IsAllowed(appId, strings.TrimSpace(p))
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
