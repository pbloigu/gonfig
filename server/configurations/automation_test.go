package configurations

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AutomationsTestSuite struct {
	suite.Suite
	repo *c
}

func TestAutomationsSuite(t *testing.T) {
	suite.Run(t, &AutomationsTestSuite{})
}

func (st *AutomationsTestSuite) SetupTest() {
	st.repo = New(st.T().TempDir() + "/tmp.sqlite").(*c)
}

func (st *AutomationsTestSuite) TestListSeriesTriggers() {

	st.repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	mt1 := SeriesTrigger{
		SeriesName: "Measurement1",
		Actions: []Action{
			{
				Description: "Action1",
				Script:      "Script1",
			},
			{
				Description: "Action2",
				Script:      "Script2",
			},
		},
	}
	mt2 := SeriesTrigger{
		SeriesName: "Measurement2",
		Actions: []Action{
			{
				Description: "Action3",
				Script:      "Script1",
			},
			{
				Description: "Action4",
				Script:      "Script2",
			},
		},
	}

	st.repo.PersistSeriesTrigger("appid1", mt1)
	st.repo.PersistSeriesTrigger("appid1", mt2)

	mts := st.repo.ListSeriesTriggers("appid1")
	st.Len(mts, 2)
	for _, mt := range mts {
		st.Len(mt.Actions, 2)
	}
	st.Len(func() []Action {
		if st, ok := st.repo.seriesTriggers.Load("appid1:Measurement1"); ok {
			return st.(SeriesTrigger).Actions
		} else {
			return []Action{}
		}
	}(), 2)
	st.Len(func() []Action {
		if st, ok := st.repo.seriesTriggers.Load("appid1:Measurement2"); ok {
			return st.(SeriesTrigger).Actions
		} else {
			return []Action{}
		}
	}(), 2)

}

func (st *AutomationsTestSuite) TestGetStatusChangeTrigger() {

	st.repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st1 := StatusChangeTrigger{
		Actions: []Action{
			{
				Description: "Action1",
				Script:      "Script1",
			},
			{
				Description: "Action2",
				Script:      "Script2",
			},
		},
	}

	st.repo.PersistStatusChangeTrigger("appid1", st1)

	tr := st.repo.GetStatusChangeTrigger("appid1")
	st.Len(tr.Actions, 2)
	st.Len(func() []Action {
		if st, ok := st.repo.statusTriggers.Load("appid1"); ok {
			return st.(StatusChangeTrigger).Actions
		} else {
			return []Action{}
		}
	}(), 2)

}

func (st *AutomationsTestSuite) TestUpdateStatusChangeTrigger() {

	st.repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st1 := StatusChangeTrigger{
		Actions: []Action{
			{
				Description: "Action1",
				Script:      "Script1",
			},
			{
				Description: "Action2",
				Script:      "Script2",
			},
		},
	}

	st.repo.PersistStatusChangeTrigger("appid1", st1)
	st.Len(func() []Action {
		if st, ok := st.repo.statusTriggers.Load("appid1"); ok {
			return st.(StatusChangeTrigger).Actions
		} else {
			return []Action{}
		}
	}(), 2)

	tr := st.repo.GetStatusChangeTrigger("appid1")
	st.Len(tr.Actions, 2)

	st2 := StatusChangeTrigger{
		Actions: []Action{
			{
				Description: "Action3",
				Script:      "Script3",
			},
			{
				Description: "Action4",
				Script:      "Script4",
			},
		},
	}
	st.repo.UpdateStatusChangeTrigger("appid1", st2)
	tr = st.repo.GetStatusChangeTrigger("appid1")
	st.Len(tr.Actions, 2)
	st.Equal("Action3", tr.Actions[0].Description)
	st.Equal("Action4", tr.Actions[1].Description)
	st.Equal("Script3", tr.Actions[0].Script)
	st.Equal("Script4", tr.Actions[1].Script)

	var cached []Action
	if st, ok := st.repo.statusTriggers.Load("appid1"); ok {
		cached = st.(StatusChangeTrigger).Actions
	} else {
		cached = []Action{}
	}

	st.Len(cached, 2)

	st.Equal("Action3", cached[0].Description)
	st.Equal("Action4", cached[1].Description)
	st.Equal("Script3", cached[0].Script)
	st.Equal("Script4", cached[1].Script)

	cached = st.repo.GetStatusChangeActions("appid1")
}
