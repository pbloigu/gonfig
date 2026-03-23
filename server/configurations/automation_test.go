package configurations

import (
	"fmt"
)

func (st *ConfigurationsTestSuite) TestListSeriesTriggers() {

	st.insertApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	id := st.insertSeriesTrigger("appid1", "Measurement1")

	for i := range 2 {
		st.insertAction(fmt.Sprintf("descr_%d_%d", id, i), fmt.Sprintf("script_%d_%d", id, i), "series_trigger_id", id)
	}

	id = st.insertSeriesTrigger("appid1", "Measurement2")
	for i := range 2 {
		st.insertAction(fmt.Sprintf("descr_%d_%d", id, i), fmt.Sprintf("script_%d_%d", id, i), "series_trigger_id", id)
	}

	mts := st.repo.ListSeriesTriggers("appid1")
	st.Len(mts, 2)
	for _, mt := range mts {
		st.Len(mt.Actions, 2)
	}
}

func (st *ConfigurationsTestSuite) TestPersistSeriesTrigger() {
	st.insertApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "app1",
	})
	tr := SeriesTrigger{
		Id:         1,
		SeriesName: "series1",
		Actions:    createDbActions(2, true),
	}
	st.repo.PersistSeriesTrigger("appid1", tr)
	_, ok := st.repo.seriesTriggers.Load(fmt.Sprintf("%s:%s", "appid1", tr.SeriesName))
	st.True(ok)

	tr = SeriesTrigger{
		Id:         1,
		SeriesName: "series2",
		Actions:    createDbActions(2, true),
	}
	st.repo.PersistSeriesTrigger("appid1", tr)
	_, ok = st.repo.seriesTriggers.Load(fmt.Sprintf("%s:%s", "appid1", tr.SeriesName))
	st.True(ok)

	trs := st.getSeriesTriggers("appid1")
	st.Len(trs, 2)
	for _, tr := range trs {
		st.Len(st.getActions("series_trigger_id", tr.Id), 2)
	}
}

func (st *ConfigurationsTestSuite) TestPersistStatusChangeTrigger() {
	st.insertApplication(Application{
		Id:     "appId",
		ApiKey: "apiKey",
		Name:   "name",
	})
	tr := StatusChangeTrigger{
		Id:      666,
		Actions: createDbActions(3, true),
	}
	st.repo.PersistStatusChangeTrigger("appId", tr)
	_, ok := st.repo.statusTriggers.Load("appId")
	st.True(ok)

	tr2 := st.getStatusChangeTrigger("appId")
	st.NotNil(tr2)
	st.Len(st.getActions("status_change_trigger_id", tr2.Id), 3)
}

func (st *ConfigurationsTestSuite) TestUpdateStatusChangeTrigger() {

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

func (st *ConfigurationsTestSuite) TestGetStatusChangeTrigger() {

	st.repo.PersistApplication(Application{
		Id:     "appid1",
		ApiKey: "apikey1",
		Name:   "appname1",
	})

	st1 := StatusChangeTrigger{
		Actions: createDbActions(2, true),
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

func (st *ConfigurationsTestSuite) TestDeleteStatusChangeTrigger() {
	st.insertApplication(Application{
		Id: "appId",
	})
	tr := StatusChangeTrigger{
		Actions: createDbActions(3, true),
	}
	id := st.insertStatusChangeTrigger("appId", tr)
	st.repo.statusTriggers.Store("appId", st)

	st.repo.DeleteStatusChangeTrigger("appId")
	st.Nil(st.getStatusChangeTrigger("appId"))
	st.Len(st.getActions("status_change_trigger_id", id), 0)
	_, ok := st.repo.statusTriggers.Load("appId")
	st.False(ok)
}

func (st *ConfigurationsTestSuite) TestGetStatusChangeActions() {

	tr := StatusChangeTrigger{
		Actions: createDbActions(3, true),
	}

	st.repo.statusTriggers.Store("appId", tr)
	st.Len(st.repo.GetStatusChangeActions("appId"), 3)
}

func (st *ConfigurationsTestSuite) TestPersistCronTrigger() {
	tr := st.repo.PersistCronTrigger(CronTrigger{
		Description:    "descr",
		CronExpression: "expr",
		Actions:        createDbActions(3, true),
	})

	dbTr := st.getCronTrigger(tr.Id)
	st.NotNil(dbTr)
	st.Equal(dbTr.Description, "descr")
	st.Equal("expr", dbTr.CronExpression)

	st.Len(st.getActions("cron_trigger_id", dbTr.Id), 3)
}

func (st *ConfigurationsTestSuite) TestGetCronTrigger() {
	id := st.insertCronTrigger(CronTrigger{
		Id:             123,
		Description:    "descr",
		CronExpression: "expr",
		Actions:        createDbActions(3, true),
	})

	tr := st.repo.GetCronTrigger(id)
	st.Equal("descr", tr.Description)
	st.Equal("expr", tr.CronExpression)
	st.Len(tr.Actions, 3)
}

func (st *ConfigurationsTestSuite) TestDeleteCronTrigger() {
	id := st.insertCronTrigger(CronTrigger{
		Id:             123,
		Description:    "descr",
		CronExpression: "expr",
		Actions:        createDbActions(3, true),
	})
	st.repo.DeleteCronTrigger(id)

	st.Nil(st.getCronTrigger(id))
	st.Len(st.getActions("cron_trigger_id", id), 0)
}

func (st *ConfigurationsTestSuite) TestUpdateCronTrigger() {
	id := st.insertCronTrigger(CronTrigger{
		Id:             123,
		Description:    "descr",
		CronExpression: "expr",
		Actions:        createDbActions(3, true),
	})

	actions := st.getActions("cron_trigger_id", id)

	st.repo.UpdateCronTrigger(CronTrigger{
		Id:             id,
		Description:    "descr2",
		CronExpression: "expr2",
		Actions:        createDbActions(2, true),
	})

	dbTr := st.getCronTrigger(id)
	st.Equal("descr2", dbTr.Description)
	st.Equal("expr2", dbTr.CronExpression)

	dbActions := st.getActions("cron_trigger_id", id)
	st.Len(dbActions, 2)

	for _, dba := range dbActions {
		st.False(func() bool {
			for _, a := range actions {
				if a.Id == dba.Id {
					return true
				}
			}
			return false
		}())
	}
}

func (st *ConfigurationsTestSuite) TestListCronTriggers() {
	for i := range 3 {
		st.insertCronTrigger(CronTrigger{
			Description:    fmt.Sprintf("descr %d", i),
			CronExpression: fmt.Sprintf("expr %d", i),
			Actions:        createDbActions(3, true),
		})
	}
	trs := st.repo.ListCronTriggers()
	st.Len(trs, 3)
	for _, tr := range trs {
		st.Len(tr.Actions, 3)
		st.True(func() bool {
			for i := range 3 {
				if tr.CronExpression == fmt.Sprintf("expr %d", i) &&
					tr.Description == fmt.Sprintf("descr %d", i) {
					return true
				}
			}
			return false
		}())
	}
}

func (st *ConfigurationsTestSuite) getCronTrigger(id int) *CronTrigger {
	r, err := st.repo.db.Context().Query(`SELECT id, description, expression FROM CronTrigger WHERE id = ?`, id)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()

	if r.Next() {
		tr := CronTrigger{}
		err = r.Scan(&tr.Id, &tr.Description, &tr.CronExpression)
		if err != nil {
			st.Fail(err.Error())
		}
		return &tr
	} else {
		return nil
	}

}

func (st *ConfigurationsTestSuite) insertSeriesTrigger(appId, seriesName string) int {
	_, err := st.repo.db.Context().Exec(`INSERT INTO SeriesTrigger (
		application_id,
		series_name
	) VALUES (?, ?)`, appId, seriesName)
	if err != nil {
		st.Fail(err.Error())
	}
	r, err := st.repo.db.Context().Query(`SELECT id FROM SeriesTrigger WHERE 
	application_id = ? AND series_name = ?`, appId, seriesName)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	var id int
	if r.Next() {
		r.Scan(&id)
		return id
	}
	st.Fail("Series not peristed.")
	return -1
}

func (st *ConfigurationsTestSuite) insertCronTrigger(tr CronTrigger) int {
	_, err := st.repo.db.Context().Exec(`INSERT INTO CronTrigger (description, expression) values (?, ?)`,
		tr.Description, tr.CronExpression)
	if err != nil {
		st.Fail(err.Error())
	}
	r, err := st.repo.db.Context().Query(`SELECT last_insert_rowid()`)
	if err != nil {
		st.Fail(err.Error())
	}
	if !r.Next() {
		st.Fail("Not inserted.")
	}
	defer r.Close()
	var id int
	r.Scan(&id)
	r.Close()
	for _, a := range tr.Actions {
		_, err := st.repo.db.Context().Exec(`INSERT INTO Action (description, 
		script, cron_trigger_id) VALUES (?,?,?)`, a.Description, a.Script, id)
		if err != nil {
			st.Fail(err.Error())
		}
	}
	return id
}

func (st *ConfigurationsTestSuite) insertStatusChangeTrigger(appId string, tr StatusChangeTrigger) int {
	_, err := st.repo.db.Context().Exec(`INSERT INTO StatusTrigger (application_id) values (?)`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	r, err := st.repo.db.Context().Query(`SELECT id FROM StatusTrigger where application_id = ?`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	if !r.Next() {
		st.Fail("Not inserted.")
	}
	defer r.Close()
	var id int
	r.Scan(&id)
	r.Close()
	for _, a := range tr.Actions {
		_, err := st.repo.db.Context().Exec(`INSERT INTO Action (description, 
		script, status_change_trigger_id) VALUES (?,?,?)`, a.Description, a.Script, id)
		if err != nil {
			st.Fail(err.Error())
		}
	}
	return id
}

func (st *ConfigurationsTestSuite) insertAction(descr, script string, triggerFkField string, trggerId int) {
	_, err := st.repo.db.Context().Exec(fmt.Sprintf(`INSERT INTO Action (
		description,
		script,
		%s)
	VALUES (?, ?, ?)`, triggerFkField), descr, script, trggerId)
	if err != nil {
		st.Fail(err.Error())
	}
}

func (st *ConfigurationsTestSuite) getStatusChangeTrigger(appId string) *StatusChangeTrigger {
	r, err := st.repo.db.Context().Query(`SELECT id FROM StatusTrigger
	WHERE application_id = ?`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	if r.Next() {
		tr := StatusChangeTrigger{}
		r.Scan(&tr.Id)
		return &tr
	} else {
		return nil
	}
}

func (st *ConfigurationsTestSuite) getSeriesTriggers(appId string) []SeriesTrigger {
	r, err := st.repo.db.Context().Query(`SELECT id, series_name FROM SeriesTrigger
	WHERE application_id = ?`, appId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	trs := make([]SeriesTrigger, 0)
	for r.Next() {
		tr := SeriesTrigger{}
		r.Scan(&tr.Id, &tr.SeriesName)
		trs = append(trs, tr)
	}
	return trs
}

func (st *ConfigurationsTestSuite) getActions(triggerFkField string, trggerId int) []Action {
	r, err := st.repo.db.Context().Query(fmt.Sprintf(`SELECT id, description, script FROM Action
	WHERE %s=?`, triggerFkField), trggerId)
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	acts := make([]Action, 0)
	for r.Next() {
		a := Action{}
		r.Scan(&a.Id, &a.Description, &a.Script)
		acts = append(acts, a)
	}
	return acts
}

func createDbActions(number int, withId bool) []Action {
	dbActions := make([]Action, 0)
	for i := range number {
		dbActions = append(dbActions, Action{
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
