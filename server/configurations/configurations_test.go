package configurations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersistApplication(t *testing.T) {
	repo := New(t.TempDir() + "/test.sqlite")
	a := repo.PersistApplication(Application{
		Id:     "appid",
		ApiKey: "apikey",
		Name:   "appname",
		Configuration: Configuration{
			Data: "some data",
		},
	})

	assert.Equal(t, "appid", a.Id, "Id was not persisted.")
	assert.Equal(t, "apikey", a.ApiKey, "API Key was not persisted.")
	assert.Equal(t, "appname", a.Name, "Name was not persisted.")
	assert.Equal(t, "some data", a.Configuration.Data, "Configuration data was not persisted.")

	assert.NotNil(t, a.Configuration.CreatedAt, "Configuration creation timestamp was nil.")

	a = repo.GetApplication(a.Id)

	assert.Equal(t, "", a.ApiKey, "API Key was returned although shouldn't have.")
	assert.Equal(t, "appname", a.Name, "Name was not returned.")
	assert.Equal(t, "some data", a.Configuration.Data, "Configuration data was not returned.")

	assert.Equal(t, "appid", a.Id, "Id was not returned.")
	assert.Equal(t, "some data", a.Configuration.Data, "Configuration data was not returned.")
	assert.NotNil(t, a.Configuration.CreatedAt, "Configuration creation timestamp was not returned.")
}

func TestListApplications(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	for i := range 10 {
		repo.PersistApplication(Application{
			Id:     fmt.Sprintf("appid_%d", i),
			ApiKey: fmt.Sprintf("apikey_%d", i),
			Name:   fmt.Sprintf("appname_%d", i),
			Configuration: Configuration{
				Data: fmt.Sprintf("somedata_%d", i),
			},
		})
	}

	apps := repo.ListApplications()

	assert.Len(t, apps, 10, "Not all apps loaded.")

	for i := range 10 {
		a := apps[i]
		assert.Equal(t, fmt.Sprintf("appid_%d", i), a.Id, "Id not what expected.")
		assert.Equal(t, fmt.Sprintf("appname_%d", i), a.Name, "Name not what expected.")
		assert.Equal(t, fmt.Sprintf("somedata_%d", i), a.Configuration.Data, "Configuration data not what expected.")
		assert.Equal(t, "", a.ApiKey, "API Key was returned although shouldn't have.")
	}
}

func TestGetConfiguration(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	a := repo.PersistApplication(Application{
		Id:     "appid",
		ApiKey: "apikey",
		Name:   "appname",
		Configuration: Configuration{
			Data: "some data",
		},
	})

	repo.PersistConfiguration(a.Id, Configuration{Data: "some data 2"})

	c := repo.GetConfiguration(a.Id)

	assert.Equal(t, "some data 2", c.Data, "Latest configuration not returned.")
}

func TestUpdateApplication(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	a := repo.PersistApplication(Application{
		Id:     "appid",
		ApiKey: "apikey",
		Name:   "appname",
		Configuration: Configuration{
			Data: "some data",
		},
	})

	repo.UpdateApplication(Application{
		Id:     a.Id,
		ApiKey: "updatedKey",
		Name:   "updatedName",
		Configuration: Configuration{
			Data: "updatedData",
		},
	})

	a = repo.GetApplication(a.Id)

	assert.Equal(t, "appid", a.Id, "Application id updated while shouldn't have.")
	assert.Equal(t, "updatedData", a.Configuration.Data, "Configuration data not updated.")
	assert.Equal(t, "updatedName", a.Name, "Application name not updated.")
}

func TestLogin(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")
	cc := repo.(*c)

	cc.db.Context().Exec("INSERT INTO User (login, password) values ('user1', 'pwd1')")
	cc.db.Context().Exec("INSERT INTO User (login, password) values ('user2', 'pwd2)")

	assert.True(t, repo.Login("user1", "pwd1"), "User not correctly logged in.")
	assert.False(t, repo.Login("user1", "pwd2"), "User incorrectly logged in.")
}

func TestIsAllowed(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	repo.PersistApplication(Application{
		Id:     "appid2",
		ApiKey: "apikey2",
		Name:   "appname2",
	})

	assert.True(t, repo.IsAllowed("appid1", "apikey1"), "Application not allowed.")
	assert.False(t, repo.Login("appid2", "apikey1"), "Application incorrectly allowed.")
}

func TestListMeasurementTriggers(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	mt1 := MeasurementTrigger{
		MeasurementName: "Measurement1",
		Actions: []Action{
			{
				Name:   "Action1",
				Script: "Script1",
			},
			{
				Name:   "Action2",
				Script: "Script2",
			},
		},
	}
	mt2 := MeasurementTrigger{
		MeasurementName: "Measurement2",
		Actions: []Action{
			{
				Name:   "Action3",
				Script: "Script1",
			},
			{
				Name:   "Action4",
				Script: "Script2",
			},
		},
	}

	repo.PersitMeasurementTrigger("appid1", mt1)
	repo.PersitMeasurementTrigger("appid1", mt2)

	mts := repo.ListMeasurementTriggers("appid1")
	assert.Len(t, mts, 2)
	for _, mt := range mts {
		assert.Len(t, mt.Actions, 2)
	}
}

func TestListStatusChangeTriggers(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st1 := StatusChangeTrigger{
		Actions: []Action{
			{
				Name:   "Action1",
				Script: "Script1",
			},
			{
				Name:   "Action2",
				Script: "Script2",
			},
		},
	}

	repo.PersistStatusChangeTrigger("appid1", st1)

	tr := repo.GetStatusChangeTrigger("appid1")
	assert.Len(t, tr.Actions, 2)

}

func TestUpdateStatusChangeTrigger(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st1 := StatusChangeTrigger{
		Actions: []Action{
			{
				Name:   "Action1",
				Script: "Script1",
			},
			{
				Name:   "Action2",
				Script: "Script2",
			},
		},
	}

	repo.PersistStatusChangeTrigger("appid1", st1)

	tr := repo.GetStatusChangeTrigger("appid1")
	assert.Len(t, tr.Actions, 2)

	st2 := StatusChangeTrigger{
		Actions: []Action{
			{
				Name:   "Action1",
				Script: "Script1",
			},
			{
				Name:   "Action2",
				Script: "Script2",
			},
		},
	}
	repo.UpdateStatusChangeTrigger("appid1", st2)
	tr = repo.GetStatusChangeTrigger("appid1")
	assert.Len(t, tr.Actions, 2)
	assert.Equal(t, "Action1", tr.Actions[0].Name)
	assert.Equal(t, "Action2", tr.Actions[1].Name)
	assert.Equal(t, "Script1", tr.Actions[0].Script)
	assert.Equal(t, "Script2", tr.Actions[1].Script)

}
