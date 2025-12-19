package automation

import (
	"errors"

	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/cc"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/rs/zerolog/log"
)

type data struct {
	s service.Service
}

type ipc struct {
	s service.Service
	r cc.Router
}

type statusCtx struct {
	ApplicationId string
	Status        string
}

type logging struct {
}

type ipcResult struct {
	List  []any
	Named map[string]any
}

func (d data) LastSeries(appId string, seriesName string) api.Series {
	return d.s.GetSeries(appId, seriesName)
}

func (d data) ListApplications() []string {
	return d.s.Cached().ListApplicationIds()
}

func asError(a any) error {
	err, ok := a.(error)
	if ok {
		return err
	} else {
		return nil
	}

}
func (l logging) Info(msg any) {
	err := asError(msg)
	if err != nil {
		log.Info().AnErr("error", err).Send()
	} else {
		log.Info().Any("message", msg).Send()
	}

}
func (l logging) Debug(msg any) {
	err := asError(msg)
	if err != nil {
		log.Debug().AnErr("error", err).Send()
	} else {
		log.Debug().Any("message", msg).Send()
	}
}

func (i ipc) Call(appId string, proc string, args []any) (ipcResult, error) {
	list, named, err := i.r.CallIpc(appId, proc, args)
	if err != nil {
		log.Debug().AnErr("error", err).Msg("IPC call failed.")
		return ipcResult{}, errors.New(err.Error())
	} else {
		return ipcResult{
			List:  list,
			Named: named,
		}, nil
	}
}
