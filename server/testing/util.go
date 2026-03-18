package testing

import (
	"fmt"

	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
)

func createDbActions(number int, withId bool) []configurations.Action {
	dbActions := make([]configurations.Action, 0)
	for i := range number {
		dbActions = append(dbActions, configurations.Action{
			Id: func() int {
				if withId {
					return i
				} else {
					return 0
				}
			}(),
			Description: fmt.Sprintf("descr: %d", i),
			Script:      fmt.Sprintf("script %d", i),
		})
	}
	return dbActions
}

func createApiActions(number int, withId bool) []api.Action {
	apiActions := make([]api.Action, 0)
	for i := range number {
		apiActions = append(apiActions, api.Action{
			Id: func() int {
				if withId {
					return i
				} else {
					return 0
				}
			}(),
			Description: fmt.Sprintf("descr: %d", i),
			Script:      fmt.Sprintf("script %d", i),
		})
	}
	return apiActions
}
