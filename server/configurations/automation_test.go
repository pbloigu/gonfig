package configurations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSeriesTriggers(t *testing.T) {
	repo := New(t.TempDir() + "/tmp.sqlite")

	repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	mt1 := SeriesTrigger{
		SeriesName: "Measurement1",
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
	mt2 := SeriesTrigger{
		SeriesName: "Measurement2",
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

	repo.PersistSeriesTrigger("appid1", mt1)
	repo.PersistSeriesTrigger("appid1", mt2)

	mts := repo.ListSeriesTriggers("appid1")
	assert.Len(t, mts, 2)
	for _, mt := range mts {
		assert.Len(t, mt.Actions, 2)
	}
	assert.Len(t, repo.(*c).seriesActions.get("appid1|Measurement1"), 2)
	assert.Len(t, repo.(*c).seriesActions.get("appid1|Measurement2"), 2)

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
	assert.Len(t, repo.(*c).statusActions.get("appid1"), 2)

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
	assert.Len(t, repo.(*c).statusActions.get("appid1"), 2)

	tr := repo.GetStatusChangeTrigger("appid1")
	assert.Len(t, tr.Actions, 2)

	st2 := StatusChangeTrigger{
		Actions: []Action{
			{
				Name:   "Action3",
				Script: "Script3",
			},
			{
				Name:   "Action4",
				Script: "Script4",
			},
		},
	}
	repo.UpdateStatusChangeTrigger("appid1", st2)
	tr = repo.GetStatusChangeTrigger("appid1")
	assert.Len(t, tr.Actions, 2)
	assert.Equal(t, "Action3", tr.Actions[0].Name)
	assert.Equal(t, "Action4", tr.Actions[1].Name)
	assert.Equal(t, "Script3", tr.Actions[0].Script)
	assert.Equal(t, "Script4", tr.Actions[1].Script)

	cached := repo.(*c).statusActions.get("appid1")
	assert.Len(t, cached, 2)

	assert.Equal(t, "Action3", cached[0].Name)
	assert.Equal(t, "Action4", cached[1].Name)
	assert.Equal(t, "Script3", cached[0].Script)
	assert.Equal(t, "Script4", cached[1].Script)

}
