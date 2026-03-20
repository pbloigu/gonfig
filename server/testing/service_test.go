package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/pbloigu/gonfig/server/series"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/pbloigu/gonfig/server/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	cMock     *MockConfigurations
	serMock   *MockSeriesDb
	schedMock *MockScheduler
	s         service.Service
}

func (st *ServiceTestSuite) SetupTest() {
	st.cMock = NewMockConfigurations(st.T())
	st.serMock = NewMockSeriesDb(st.T())
	st.schedMock = NewMockScheduler(st.T())
	st.cMock.EXPECT().ListCronTriggers().Return([]configurations.CronTrigger{}).Once()
	st.s = service.New(st.serMock, st.cMock, st.schedMock)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, &ServiceTestSuite{})
}

func (st *ServiceTestSuite) TestIsValid() {

	r := api.CronValidationRequest{
		Minute: "0",
		Hour:   "0",
		Dom:    "1",
		Month:  "1",
		Dow:    "6",
	}

	st.True(st.s.IsValid(r))

	r = api.CronValidationRequest{
		Minute: "15-59/10,*/2",
		Hour:   "*",
		Dom:    "*",
		Month:  "*",
		Dow:    "*",
	}
	st.True(st.s.IsValid(r))

	r = api.CronValidationRequest{
		Minute: "15-59/10,*/2",
		Hour:   "*",
		Dom:    "*",
		Month:  "*",
		Dow:    "*",
	}
	st.True(st.s.IsValid(r))

}

func (st *ServiceTestSuite) TestNewPagination() {

	p := st.s.NewPagination(10, 5, 3)
	st.Equal(p.Page, 3)
	st.Equal(p.Size, 10)

	p = st.s.NewPagination(0, 5, 0)
	st.Equal(p.Page, 1)
	st.Equal(p.Size, 5)

	p = st.s.NewPagination(-1, 5, 0)
	st.Equal(p.Page, 1)
	st.Equal(p.Size, 5)
}

func (st *ServiceTestSuite) TestNewSort() {

	sort := st.s.NewSort("name", "default", "ASC")
	st.Equal("name", sort.Sort)
	st.Equal(service.ASC, sort.Dir)

	sort = st.s.NewSort("", "default", "DESC")
	st.Equal("default", sort.Sort)
	st.Equal(service.DESC, sort.Dir)

	sort = st.s.NewSort("", "default", "")
	st.Equal("default", sort.Sort)
	st.Equal(service.DESC, sort.Dir)

	sort = st.s.NewSort("id", "default", "invalid")
	st.Equal("id", sort.Sort)
	st.Equal(service.DESC, sort.Dir)
}

func (st *ServiceTestSuite) TestUpdateAppication() {

	cApp := configurations.Application{
		Id:   "testid",
		Name: "testname",
		Configuration: configurations.Configuration{
			Data: "testdata",
		},
	}
	st.cMock.EXPECT().UpdateApplication(cApp).Once()
	st.cMock.EXPECT().GetApplication("testid").Once().Return(configurations.Application{}) // no too interested in the return value

	st.s.UpdateApplication(api.Application{
		Id:       "testid",
		ApiKey:   "shouldskip",
		Name:     "testname",
		Hostname: "shouldskip",
		Ip:       "shouldskip",
		Configuration: api.Configuration{
			Data: "testdata",
		},
	})
}

func (st *ServiceTestSuite) TestAddApplication() {
	ec := make(chan events.ApplicationAdded, 1)
	event.On(func(e events.ApplicationAdded) {
		ec <- e
	})
	appId := "appId"
	apiKey := "apiKey"
	st.cMock.EXPECT().PersistApplication(mock.AnythingOfType("Application")).RunAndReturn(func(a configurations.Application) configurations.Application {
		st.NotEqual("", a.ApiKey)
		st.NotEqual("", a.Id)
		st.Equal("testname", a.Name)
		st.Equal("testdata", a.Configuration.Data)
		return configurations.Application{
			Id:     appId,
			Name:   "testname",
			ApiKey: apiKey,
			Configuration: configurations.Configuration{
				Data: "testdata",
			},
		}
	}).Once()

	app := st.s.AddApplication(api.Application{
		Name: "testname",
		Configuration: api.Configuration{
			Data: "testdata",
		},
	})
	st.Equal("appId", app.Id)
	st.Equal("apiKey", app.ApiKey)
	st.Equal("testname", app.Name)
	st.Equal("testdata", app.Configuration.Data)
	a := <-ec
	st.Equal("appId", a.AppId)

}

func (st *ServiceTestSuite) TestGetApplication() {

	cApp := configurations.Application{
		Id:   "appId",
		Name: "appName",
		Configuration: configurations.Configuration{
			Data:      "testdata",
			CreatedAt: time.Now(),
		},
	}

	st.cMock.EXPECT().GetApplication("appId").Return(cApp).Once()
	app := st.s.GetApplication("appId")
	st.Equal(cApp.Id, app.Id)
	st.Equal(cApp.Name, app.Name)
	st.Equal(cApp.Configuration.Data, app.Configuration.Data)
	st.Equal(cApp.Configuration.CreatedAt, app.Configuration.Date)
	st.Equal("", app.ApiKey)
}

func (st *ServiceTestSuite) TestListApplications() {

	cApps := make([]configurations.Application, 0)
	cApps = append(cApps, configurations.Application{
		Id:   "appId1",
		Name: "appName1",
		Configuration: configurations.Configuration{
			Data:      "testData1",
			CreatedAt: time.Now().Add(time.Minute * 1),
		},
	})
	cApps = append(cApps, configurations.Application{
		Id:   "appId2",
		Name: "appName2",
		Configuration: configurations.Configuration{
			Data:      "testData2",
			CreatedAt: time.Now().Add(time.Minute * 2),
		},
	})

	st.cMock.EXPECT().ListApplications().Return(cApps).Once()

	apps := st.s.ListApplications()
	st.Equal(len(cApps), len(apps))
	for i, cApp := range cApps {
		st.Equal(cApp.Id, apps[i].Id)
		st.Equal(cApp.Name, apps[i].Name)
		st.Equal(cApp.Configuration.Data, apps[i].Configuration.Data)
		st.Equal(cApp.Configuration.CreatedAt, apps[i].Configuration.Date)
		st.Equal("", apps[i].ApiKey)
	}
}

func (st *ServiceTestSuite) TestDeleteApplication() {

	ec := make(chan events.ApplicationDeleted, 1)
	event.On(func(e events.ApplicationDeleted) {
		ec <- e
	})

	st.cMock.EXPECT().DeleteApplication("appId").Once()
	st.serMock.EXPECT().DeleteApplication("appId").Once()
	st.s.DeleteApplication("appId")
	a := <-ec
	st.Equal("appId", a.AppId)
}

func (st *ServiceTestSuite) TestAddConfiguration() {

	st.cMock.EXPECT().PersistConfiguration("appId", mock.AnythingOfType("Configuration")).Run(func(applicationId string, c configurations.Configuration) {
		st.Equal("appId", applicationId)
		st.Equal("testData", c.Data)
	}).Once()

	st.s.AddConfiguration("appId", api.Configuration{
		Data: "testData",
	})
}

func (st *ServiceTestSuite) TestGetConfiguration() {

	cConf := configurations.Configuration{
		CreatedAt: time.Now(),
		Data:      "testData",
	}
	st.cMock.EXPECT().GetConfiguration("appId").Return(cConf).Once()
	conf := st.s.GetConfiguration("appId")
	st.Equal(cConf.Data, conf.Data)
	st.Equal(cConf.CreatedAt, conf.Date)
}

func (st *ServiceTestSuite) TestAddSeriesValue() {

	recdAt := int(time.Now().Unix())
	data := "someData"
	serValue := series.SeriesValue{
		RecordedAt: &recdAt,
		Data:       &data,
	}

	st.serMock.EXPECT().PeristSeriesValue("appId", "seriesName", serValue).Once()
	st.s.AddSeriesValue("appId", "seriesName", api.SeriesValue{
		Recoded: &recdAt,
		Data:    &data,
	})
}

func (st *ServiceTestSuite) TestInitSeries() {

	st.serMock.EXPECT().InitSeries("appId", "seriesName").Once()
	st.s.InitSeries("appId", api.Series{
		Name: "seriesName",
	})
}

func (st *ServiceTestSuite) TestGetSeries() {
	lValueTime := 123
	lRecorded := 456
	lValue := "lastValue"
	dbSeries := series.Series{
		Id:                666,
		LastValueTime:     &lValueTime,
		LastValueRecorded: &lRecorded,
		LastValue:         &lValue,
		Name:              "seriesName",
	}
	st.serMock.EXPECT().GetSeries("appId", "seriesName").Return(dbSeries).Once()
	series := st.s.GetSeries("appId", "seriesName")
	st.Equal(dbSeries.Name, series.Name)
	st.Equal(*dbSeries.LastValue, *series.LastValue)
	st.Equal(*dbSeries.LastValueTime, *series.LastValueTime)
	st.Equal(*dbSeries.LastValueRecorded, *series.LastValueRecorded)
}

func (st *ServiceTestSuite) TestListSeriesValues() {
	lValueTime := 123
	lRecorded := 456
	lValue := "lastValue"
	dbSeries := series.Series{
		Id:                666,
		LastValueTime:     &lValueTime,
		LastValueRecorded: &lRecorded,
		LastValue:         &lValue,
		Name:              "seriesName",
	}
	dbValues := make([]series.SeriesValue, 0)
	for i := range 2 {
		recd := i + 1
		data := fmt.Sprintf("%d", i)
		dbValues = append(dbValues, series.SeriesValue{
			CreatedAt:  i,
			RecordedAt: &recd,
			Data:       &data,
		})
	}
	st.serMock.EXPECT().ListSeriesValues(666, series.Sort{Field: series.CREATED, Dir: series.ASC}, series.Pagination{Page: 2, Size: 10}).Return(dbValues).Once()
	st.serMock.EXPECT().GetSeries("appId", "seriesName").Return(dbSeries).Once()
	st.serMock.EXPECT().CountSeriesValues(666).Return(100).Once()

	series := st.s.ListSeriesValues("appId", "seriesName", service.Sort{Sort: "created", Dir: service.ASC}, service.Pagination{Page: 2, Size: 10})
	st.Equal("seriesName", series.Series.Name)
	st.Equal(*series.Series.LastValue, lValue)
	st.Equal(*series.Series.LastValueRecorded, lRecorded)
	st.Equal(*series.Series.LastValueTime, lValueTime)

	st.Equal(100, series.Total)
	st.Equal(2, series.Page)
	st.Equal(10, series.PageSize)

	st.Equal(len(dbValues), len(series.Values))

	for i, v := range series.Values {
		st.Equal(dbValues[i].CreatedAt, v.Time)
		st.Equal(*dbValues[i].RecordedAt, *v.Recoded)
		st.Equal(*dbValues[i].Data, *v.Data)
	}
}

func (st *ServiceTestSuite) TestListSeries() {
	dbSeries := make([]series.Series, 0)
	for i := range 2 {
		lValueTime := i + 2
		lRecd := i + 3
		lValue := fmt.Sprintf("%d", i)
		dbSeries = append(dbSeries, series.Series{
			Id:                i,
			LastValueTime:     &lValueTime,
			LastValueRecorded: &lRecd,
			LastValue:         &lValue,
			Name:              fmt.Sprintf("series %d", i),
		})
	}
	st.serMock.EXPECT().ListSeries("appId").Return(dbSeries).Once()

	series := st.s.ListSeries("appId")

	st.Equal(len(dbSeries), len(series))

	for i, s := range series {
		st.Equal(dbSeries[i].Name, s.Name)
		st.Equal(*dbSeries[i].LastValue, *s.LastValue)
		st.Equal(*dbSeries[i].LastValueRecorded, *s.LastValueRecorded)
		st.Equal(*dbSeries[i].LastValueTime, *s.LastValueTime)
	}
}

func (st *ServiceTestSuite) TestIsAllowed() {
	st.cMock.EXPECT().IsAllowed("appId", "apiKey").Return(false).Once()
	isAllowed := st.s.IsAllowed("appId", "apiKey")
	st.Equal(false, isAllowed)
}

func (st *ServiceTestSuite) TestLogin() {
	st.cMock.EXPECT().Login("login", "password").Return(false).Once()
	login := st.s.Login("login", "password")
	st.Equal(false, login)
}

func (st *ServiceTestSuite) TestAddStatusChangeTrigger() {
	apiActions := createApiActions(5, true)
	apiTrigger := api.StatusChangeTrigger{
		Id:      123,
		Actions: apiActions,
	}

	dbTrigger := util.StatusChangeTriggerFromApi(apiTrigger)

	dbActions2 := make([]configurations.Action, 0)
	for i, a := range apiActions {
		dbActions2 = append(dbActions2, configurations.Action{
			Id:          i,
			Description: a.Description,
			Script:      a.Script,
		})
	}
	dbTrigger2 := configurations.StatusChangeTrigger{
		Id:      1,
		Actions: dbActions2,
	}

	st.cMock.EXPECT().PersistStatusChangeTrigger("appId", dbTrigger).Return(dbTrigger2).Once()
	apiTrigger2 := st.s.AddStatusChangeTrigger("appId", apiTrigger)

	st.Equal(dbTrigger2.Id, apiTrigger2.Id)
	st.Equal(len(dbActions2), len(apiTrigger2.Actions))

	for i, v := range apiTrigger2.Actions {
		st.Equal(dbActions2[i].Id, v.Id)
		st.Equal(dbActions2[i].Description, v.Description)
		st.Equal(dbActions2[i].Script, v.Script)
	}
}

func (st *ServiceTestSuite) TestDeleteStatusChangeTrigger() {
	st.cMock.EXPECT().DeleteStatusChangeTrigger("appId").Once()
	st.s.DeleteStatusChangeTrigger("appId")
}

func (st *ServiceTestSuite) TestUpdateStatusChangeTrigger() {
	apiActions := createApiActions(5, true)
	apiTrigger := api.StatusChangeTrigger{
		Id:      123,
		Actions: apiActions,
	}

	dbTrigger := util.StatusChangeTriggerFromApi(apiTrigger)

	dbActions2 := make([]configurations.Action, 0)
	for i, a := range apiActions {
		dbActions2 = append(dbActions2, configurations.Action{
			Id:          i,
			Description: a.Description,
			Script:      a.Script,
		})
	}
	dbTrigger2 := configurations.StatusChangeTrigger{
		Id:      0,
		Actions: dbActions2,
	}

	st.cMock.EXPECT().UpdateStatusChangeTrigger("appId", dbTrigger).Once()
	st.cMock.EXPECT().GetStatusChangeTrigger("appId").Return(dbTrigger2)
	apiTrigger2 := st.s.UpdateStatusChangeTrigger("appId", apiTrigger)

	st.Equal(len(dbTrigger2.Actions), len(apiTrigger2.Actions))

	for i, v := range apiTrigger2.Actions {
		st.Equal(dbActions2[i].Id, v.Id)
		st.Equal(dbActions2[i].Description, v.Description)
		st.Equal(dbActions2[i].Script, v.Script)
	}
}

func (st *ServiceTestSuite) TestGetStatusChangeTrigger() {
	dbActions := createDbActions(5, true)
	dbTrigger := configurations.StatusChangeTrigger{
		Id:      1,
		Actions: dbActions,
	}
	st.cMock.EXPECT().GetStatusChangeTrigger("appId").Return(dbTrigger).Once()
	st.cMock.EXPECT().GetStatusChangeTrigger("appId2").Return(configurations.StatusChangeTrigger{}).Once()

	apiTrigger := st.s.GetStatusChangeTrigger("appId")
	st.Equal(dbTrigger.Id, apiTrigger.Id)
	st.Equal(len(dbTrigger.Actions), len(apiTrigger.Actions))
	for i, v := range apiTrigger.Actions {
		st.Equal(dbActions[i].Id, v.Id)
		st.Equal(dbActions[i].Description, v.Description)
		st.Equal(dbActions[i].Script, v.Script)
	}
	apiTrigger2 := st.s.GetStatusChangeTrigger("appId2")
	st.Nil(apiTrigger2)
}

func (st *ServiceTestSuite) TestGetCronTrigger() {
	dbActions := createDbActions(5, true)

	dbTrigger := configurations.CronTrigger{
		Id:             1,
		Description:    "Description",
		CronExpression: "Cron expression",
		Actions:        dbActions,
	}
	st.cMock.EXPECT().GetCronTrigger(1).Return(dbTrigger).Once()
	st.cMock.EXPECT().GetCronTrigger(2).Return(configurations.CronTrigger{}).Once()

	apiTrigger := st.s.GetCronTrigger(1)
	nilTrigger := st.s.GetCronTrigger(2)

	st.Equal(dbTrigger.Id, apiTrigger.Id)
	st.Equal(dbTrigger.Description, apiTrigger.Description)
	st.Equal(dbTrigger.CronExpression, apiTrigger.CronExpression)
	st.Equal(len(dbTrigger.Actions), len(apiTrigger.Actions))

	for i, v := range apiTrigger.Actions {
		st.Equal(v.Id, v.Id)
		st.Equal(dbActions[i].Description, v.Description)
		st.Equal(dbActions[i].Script, v.Script)
	}
	st.Nil(nilTrigger)
}

func (st *ServiceTestSuite) TestAddCronTrigger() {
	apiActions := createApiActions(5, true)
	apiTrigger := api.CronTrigger{
		Id:             1,
		Description:    "Description",
		CronExpression: "Expression",
		Actions:        apiActions,
	}

	dbTrigger := util.CronTriggerFromApi(apiTrigger)

	dbActions2 := createDbActions(5, true)
	dbTrigger2 := configurations.CronTrigger{
		Id:             1,
		Description:    apiTrigger.Description,
		CronExpression: apiTrigger.CronExpression,
		Actions:        dbActions2,
	}

	st.cMock.EXPECT().PersistCronTrigger(dbTrigger).Return(dbTrigger2).Once()
	st.schedMock.EXPECT().RegisterCronTrigger(util.CronTriggerToApi(dbTrigger2))
	apiTrigger2 := st.s.AddCronTrigger(apiTrigger)

	st.Equal(dbTrigger2.Id, apiTrigger2.Id)
	st.Equal(dbTrigger2.Description, apiTrigger2.Description)
	st.Equal(dbTrigger2.CronExpression, apiTrigger2.CronExpression)
	st.Equal(len(dbTrigger2.Actions), len(apiTrigger2.Actions))

	for i, v := range dbActions2 {
		st.Equal(v.Id, apiTrigger2.Actions[i].Id)
		st.Equal(v.Description, apiTrigger2.Actions[i].Description)
		st.Equal(v.Script, apiTrigger2.Actions[i].Script)
	}
}

func (st *ServiceTestSuite) TestListCronTriggers() {
	dbTriggers := []configurations.CronTrigger{{
		Id:             1,
		Description:    "descr 1",
		CronExpression: "cron 1",
		Actions:        createDbActions(3, true),
	},
		{
			Id:             2,
			Description:    "descr 2",
			CronExpression: "cron 2",
			Actions:        createDbActions(4, true),
		}}

	st.cMock.EXPECT().ListCronTriggers().Return(dbTriggers).Once()
	apiTriggers := st.s.ListCronTriggers()

	st.Equal(len(dbTriggers), len(apiTriggers))
	for i, v := range apiTriggers {
		st.Equal(dbTriggers[i].Id, v.Id)
		st.Equal(dbTriggers[i].Description, v.Description)
		st.Equal(dbTriggers[i].CronExpression, v.CronExpression)
		for j, a := range v.Actions {
			st.Equal(dbTriggers[i].Actions[j].Id, a.Id)
			st.Equal(dbTriggers[i].Actions[j].Description, a.Description)
			st.Equal(dbTriggers[i].Actions[j].Script, a.Script)
		}
	}
}

func (st *ServiceTestSuite) TestDeleteCronTrigger() {
	st.cMock.EXPECT().DeleteCronTrigger(123).Once()
	st.schedMock.EXPECT().DeregisterCronTrigger(123)
	st.s.DeleteCronTrigger(123)
}

func (st *ServiceTestSuite) TestUpdateCronTrigger() {
	apiTrigger := api.CronTrigger{
		Id:             1,
		Description:    "desrc",
		CronExpression: "cron expr",
		Actions:        createApiActions(3, true),
	}
	dbTrigger := util.CronTriggerFromApi(apiTrigger)

	st.cMock.EXPECT().UpdateCronTrigger(dbTrigger).Once()
	st.schedMock.EXPECT().DeregisterCronTrigger(apiTrigger.Id).Once()
	st.schedMock.EXPECT().RegisterCronTrigger(apiTrigger).Once()
	st.cMock.EXPECT().GetCronTrigger(apiTrigger.Id).Return(dbTrigger).Once()
	apiTrigger2 := st.s.UpdateCronTrigger(apiTrigger)

	st.Equal(dbTrigger.Id, apiTrigger2.Id)
	st.Equal(dbTrigger.Description, apiTrigger2.Description)
	st.Equal(dbTrigger.CronExpression, apiTrigger2.CronExpression)
	st.Equal(len(dbTrigger.Actions), len(apiTrigger2.Actions))

	for i, v := range apiTrigger.Actions {
		st.Equal(dbTrigger.Actions[i].Id, v.Id)
		st.Equal(dbTrigger.Actions[i].Description, v.Description)
		st.Equal(dbTrigger.Actions[i].Script, v.Script)
	}
}

func (st *ServiceTestSuite) TestJoined() {
	st.False(st.s.IsOnline("appId"))
	st.s.Joined("appId")
	st.True(st.s.IsOnline("appId"))
	st.s.Left("appId")
	st.False(st.s.IsOnline("appId"))
}
