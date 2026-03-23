package configurations

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ConfigurationsTestSuite struct {
	suite.Suite
	repo *c
}

func TestConfigurationsSuite(t *testing.T) {
	suite.Run(t, &ConfigurationsTestSuite{})
}

func (st *ConfigurationsTestSuite) SetupTest() {
	st.repo = New(st.T().TempDir() + "/tmp.sqlite").(*c)
}

func (st *ConfigurationsTestSuite) TestPersistApplication() {
	initial := Application{
		Id:       "appId",
		ApiKey:   "apiKey",
		Name:     "name",
		Hostname: "hostname",
		Ip:       "ip",
		Configuration: Configuration{
			CreatedAt: time.Now().Add(time.Hour * -1),
			Data:      "data",
		},
	}
	returned := st.repo.PersistApplication(initial)
	_, ok := st.repo.apps.Load(initial.Id)
	st.True(ok)
	realApp := st.getApplication("appId")
	realConfig := st.getConfiguration("appId")
	st.NotNil(realApp)
	st.NotNil(realConfig)

	st.Equal(initial.Id, returned.Id)
	st.Equal(initial.Id, realApp.Id)

	st.Equal(initial.ApiKey, returned.ApiKey)
	st.Equal(initial.ApiKey, realApp.ApiKey)

	st.Equal(initial.Name, returned.Name)
	st.Equal(initial.Name, realApp.Name)

	// not stored currently
	st.Equal("", returned.Hostname)
	st.Equal("", realApp.Hostname)

	// not stored currently
	st.Equal("", returned.Ip)
	st.Equal("", realApp.Ip)

	// should not be insertable
	st.NotEqual(initial.Configuration.CreatedAt, returned.Configuration.CreatedAt)
	st.NotEqual(initial.Configuration.CreatedAt, realConfig.CreatedAt)
	st.Equal(returned.Configuration.CreatedAt, realConfig.CreatedAt)

	st.Equal(initial.Configuration.Data, returned.Configuration.Data)
	st.Equal(initial.Configuration.Data, returned.Configuration.Data)
}

func (st *ConfigurationsTestSuite) TestDeleteApplication() {
	st.insertApplication(Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "appName",
	})
	st.insertConfiguration("appId", false, Configuration{
		Data: "foodata",
	})
	st.insertConfiguration("appId", true, Configuration{
		Data: "foodata",
	})
	st.repo.apps.Store("appId", true)
	st.repo.DeleteApplication("appId")
	_, ok := st.repo.apps.Load("appId")
	st.False(ok)
	st.Nil(st.getApplication("appId"))
	st.Equal(0, len(st.getConfigurations("appId")))
}

func (st *ConfigurationsTestSuite) TestGetApplication() {
	expected := Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "appName",
	}
	st.insertApplication(expected)
	st.insertConfiguration("appId", false, Configuration{
		Data: "not latest",
	})
	st.insertConfiguration("appId", true, Configuration{
		Data: "latest",
	})
	app := st.repo.GetApplication("appId")
	st.Equal("", app.ApiKey)
	st.Equal(expected.Id, app.Id)
	st.Equal(expected.Name, app.Name)
	st.Equal("latest", app.Configuration.Data)
}

func (st *ConfigurationsTestSuite) TestListApplications() {
	for i := range 3 {
		st.insertApplication(Application{
			Id:     fmt.Sprintf("appid_%d", i),
			ApiKey: fmt.Sprintf("apikey_%d", i),
			Name:   fmt.Sprintf("app_%d", i),
		})
		for j := range 3 {
			st.insertConfiguration(fmt.Sprintf("appid_%d", i), j == 2, Configuration{
				Data: fmt.Sprintf("data_%d", j),
			})
		}
	}
	apps := st.repo.ListApplications()
	st.Len(apps, 3)
	// XXX: strong assumption about the ordering of the results
	for i := range 3 {
		app := apps[i]
		st.Equal(fmt.Sprintf("appid_%d", i), app.Id)
		st.Equal("", app.ApiKey)
		st.Equal(fmt.Sprintf("app_%d", i), app.Name)
		st.Equal("data_2", app.Configuration.Data)
	}
}

func (st *ConfigurationsTestSuite) TestUpdateApplication() {
	st.insertApplication(Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "name",
	})
	st.insertConfiguration("appId", true, Configuration{
		Data: "data",
	})
	st.repo.UpdateApplication(Application{
		Id:   "appId",
		Name: "name2",
		Configuration: Configuration{
			Data: "data2",
		},
	})
	app := st.getApplication("appId")
	st.Equal("name2", app.Name)
	config := st.getConfiguration("appId")
	st.Equal("data2", config.Data)
}

func (st *ConfigurationsTestSuite) TestListApplicationIds() {
	for i := range 5 {
		st.repo.apps.Store(fmt.Sprintf("appid_%d", i), true)
	}
	ids := st.repo.ListApplicationIds()
	st.Len(ids, 5)
	for i := range 5 {
		found := false
		for _, id := range ids {
			if found = id == fmt.Sprintf("appid_%d", i); found {
				break
			}
		}
		st.True(found)
	}
}
func (st *ConfigurationsTestSuite) TestPersistConfiguration() {
	st.insertApplication(Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "name",
	})
	st.insertConfiguration("appId", true, Configuration{
		Data: "config1",
	})
	st.repo.PersistConfiguration("appId", Configuration{
		Data: "config2",
	})
	configs := st.getConfigurations("appId")
	st.Len(configs, 2)
	config := st.getConfiguration("appId")
	st.Equal("config2", config.Data)
}

func (st *ConfigurationsTestSuite) TestGetConfiguration() {
	st.insertApplication(Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "name",
	})
	st.insertConfiguration("appId", false, Configuration{
		Data: "config1",
	})
	st.insertConfiguration("appId", true, Configuration{
		Data: "config2",
	})
	config := st.repo.GetConfiguration("appId")
	st.Equal("config2", config.Data)

}

func (st *ConfigurationsTestSuite) TestIsAllowed() {

	st.insertApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st.insertApplication(Application{
		Id:     "appid2",
		ApiKey: "apikey2",
		Name:   "appname2",
	})

	st.True(st.repo.IsAllowed("appid1", "apikey1"), "Application not allowed.")
	st.False(st.repo.Login("appid2", "apikey1"), "Application incorrectly allowed.")
}

func (st *ConfigurationsTestSuite) TestLogin() {

	st.insertUser("user1", "pwd1")
	st.insertUser("user2", "pwd2")

	st.True(st.repo.Login("user1", "pwd1"), "User not correctly logged in.")
	st.False(st.repo.Login("user1", "pwd2"), "User incorrectly logged in.")
}

func (st *ConfigurationsTestSuite) getApplication(appId string) *Application {
	r, err := st.repo.db.Context().Query(`SELECT id, name, api_key, hostname, ip FROM Application WHERE id = ?`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()

	if r.Next() {
		app := Application{}
		r.Scan(&app.Id, &app.Name, &app.ApiKey, &app.Hostname, &app.Ip)
		return &app
	}
	return nil
}

func (st *ConfigurationsTestSuite) insertApplication(app Application) {
	_, err := st.repo.db.Context().Exec(`
		INSERT INTO Application (id, name, api_key) VALUES (?, ?, ?)
	`, app.Id, app.Name, app.ApiKey)
	if err != nil {
		st.Fail(err.Error())
	}
}

func (st *ConfigurationsTestSuite) insertConfiguration(appId string, isLatest bool, config Configuration) {
	_, err := st.repo.db.Context().Exec(`
		INSERT INTO Configuration (application_id, is_latest, data) VALUES (?, ?, ?)
	`, appId, isLatest, config.Data)
	if err != nil {
		st.Fail(err.Error())
	}
}

func (st *ConfigurationsTestSuite) getConfiguration(appId string) *Configuration {
	r, err := st.repo.db.Context().Query(`SELECT created, data FROM Configuration 
										WHERE application_id = ? and is_latest = 1`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()

	if r.Next() {
		config := Configuration{}
		r.Scan(&config.CreatedAt, &config.Data)
		st.False(r.Next())
		return &config
	}
	return nil
}

func (st *ConfigurationsTestSuite) getConfigurations(appId string) []Configuration {
	r, err := st.repo.db.Context().Query(`SELECT created, data FROM Configuration 
										WHERE application_id = ?`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	configs := make([]Configuration, 0)
	for r.Next() {
		c := Configuration{}
		r.Scan(&c.CreatedAt, &c.Data)
		configs = append(configs, c)
	}
	return configs
}

func (st *ConfigurationsTestSuite) insertUser(login, password string) {
	_, err := st.repo.db.Context().Exec(`INSERT INTO User (login, password) values (?, ?)`, login, password)
	if err != nil {
		st.Fail(err.Error())
	}
}
