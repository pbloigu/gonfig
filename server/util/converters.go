package util

import (
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
)

func ActionToApi(action configurations.Action) api.Action {
	return api.Action{
		Id:          action.Id,
		Description: action.Description,
		Script:      action.Script,
	}
}

func ActionsToApi(actions []configurations.Action) []api.Action {
	acts := make([]api.Action, 0)
	for _, a := range actions {
		acts = append(acts, api.Action{
			Id:          a.Id,
			Description: a.Description,
			Script:      a.Script,
		})
	}
	return acts
}

func ActionFromApi(action api.Action) configurations.Action {
	return configurations.Action{
		Id:          action.Id,
		Description: action.Description,
		Script:      action.Script,
	}
}

func ActionsFromApi(actions []api.Action) []configurations.Action {
	acts := make([]configurations.Action, 0)
	for _, a := range actions {
		acts = append(acts, configurations.Action{
			Id:          a.Id,
			Description: a.Description,
			Script:      a.Script,
		})
	}
	return acts
}

func StatusChageTriggerToApi(trigger configurations.StatusChangeTrigger) api.StatusChangeTrigger {
	return api.StatusChangeTrigger{
		Id:      trigger.Id,
		Actions: ActionsToApi(trigger.Actions),
	}
}

func StatusChangeTriggerFromApi(trigger api.StatusChangeTrigger) configurations.StatusChangeTrigger {
	return configurations.StatusChangeTrigger{
		Id:      trigger.Id,
		Actions: ActionsFromApi(trigger.Actions),
	}
}

func CronTriggerToApi(trigger configurations.CronTrigger) api.CronTrigger {
	return api.CronTrigger{
		Id:             trigger.Id,
		Description:    trigger.Description,
		CronExpression: trigger.CronExpression,
		Actions:        ActionsToApi(trigger.Actions),
	}
}

func CronTriggerFromApi(trigger api.CronTrigger) configurations.CronTrigger {
	return configurations.CronTrigger{
		Id:             trigger.Id,
		Description:    trigger.Description,
		CronExpression: trigger.CronExpression,
		Actions:        ActionsFromApi(trigger.Actions),
	}
}
