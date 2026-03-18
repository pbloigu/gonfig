package util

import (
	"testing"

	"github.com/pbloigu/gonfig/server/configurations"

	"github.com/stretchr/testify/require"
)

func TestActionConversion(t *testing.T) {
	in := configurations.Action{Id: 1, Description: "d", Script: "s"}
	apiAction := ActionToApi(in)
	require.EqualValues(t, in.Id, apiAction.Id)
	require.Equal(t, in.Description, apiAction.Description)
	require.Equal(t, in.Script, apiAction.Script)

	out := ActionFromApi(apiAction)
	require.EqualValues(t, in, out)
}

func TestActionsConversion(t *testing.T) {
	in := []configurations.Action{{Id: 1, Description: "d1", Script: "s1"}, {Id: 2, Description: "d2", Script: "s2"}}
	apiActions := ActionsToApi(in)
	require.Len(t, apiActions, 2)
	for i := range in {
		require.EqualValues(t, in[i].Id, apiActions[i].Id)
		require.Equal(t, in[i].Description, apiActions[i].Description)
		require.Equal(t, in[i].Script, apiActions[i].Script)
	}

	out := ActionsFromApi(apiActions)
	require.Equal(t, in, out)
}

func TestStatusChangeTriggerConversion(t *testing.T) {
	in := configurations.StatusChangeTrigger{
		Id:      2,
		Actions: []configurations.Action{{Id: 10, Description: "act", Script: "echo 1"}},
	}
	apiTrigger := StatusChageTriggerToApi(in)
	require.EqualValues(t, in.Id, apiTrigger.Id)
	require.Len(t, apiTrigger.Actions, 1)
	require.EqualValues(t, in.Actions[0].Id, apiTrigger.Actions[0].Id)
	require.Equal(t, in.Actions[0].Description, apiTrigger.Actions[0].Description)
	require.Equal(t, in.Actions[0].Script, apiTrigger.Actions[0].Script)

	out := StatusChangeTriggerFromApi(apiTrigger)
	require.Equal(t, in.Id, out.Id)
	require.Equal(t, in.Actions, out.Actions)
}

func TestCronTriggerConversion(t *testing.T) {
	in := configurations.CronTrigger{
		Id:             3,
		Description:    "cron",
		CronExpression: "0 0 * * *",
		Actions:        []configurations.Action{{Id: 20, Description: "cron-act", Script: "echo cron"}},
	}
	apiTrigger := CronTriggerToApi(in)
	require.EqualValues(t, in.Id, apiTrigger.Id)
	require.Equal(t, in.Description, apiTrigger.Description)
	require.Equal(t, in.CronExpression, apiTrigger.CronExpression)
	require.Len(t, apiTrigger.Actions, 1)
	// action inner fields already implicitly tested by ActionsToApi
	require.EqualValues(t, in.Actions[0].Id, apiTrigger.Actions[0].Id)
	require.Equal(t, in.Actions[0].Description, apiTrigger.Actions[0].Description)
	require.Equal(t, in.Actions[0].Script, apiTrigger.Actions[0].Script)

	out := CronTriggerFromApi(apiTrigger)
	require.Equal(t, in.Id, out.Id)
	require.Equal(t, in.Description, out.Description)
	require.Equal(t, in.CronExpression, out.CronExpression)
	require.Equal(t, in.Actions, out.Actions)
}
