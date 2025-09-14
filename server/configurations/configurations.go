package configurations

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

type Configurations interface {
	DeleteApplication(id string)
	GetApplication(id string) Application
	GetConfiguration(appId string) Configuration
	IsAllowed(id string, apiKey string) bool
	ListApplications() []Application
	Login(login string, password string) bool
	PersistApplication(a Application) Application
	PersistConfiguration(applicationId string, c Configuration)
	UpdateApplication(a Application)
	ListMeasurementTriggers(appId string) []MeasurementTrigger
	PersitMeasurementTrigger(appId string, mt MeasurementTrigger)
	PersistStatusChangeTrigger(appId string, st StatusChangeTrigger) StatusChangeTrigger
	UpdateStatusChangeTrigger(appId string, t StatusChangeTrigger)
	GetStatusChangeTrigger(appId string) StatusChangeTrigger
	DeleteStatusChangeTrigger(appId string)
	GetStatusChangeActions(appId string) []Action
	ListApplicationIds() []string
}

type c struct {
	db                 database.Database
	statusActions      actionCache
	cronActions        actionCache
	measurementActions actionCache
	apps               appCache
}

//go:embed schema.sql
var ddl string

func New(dbLoc string) Configurations {
	c := &c{
		db: startDatabase(dbLoc),
		statusActions: actionCache{
			c: make(map[string][]Action),
			m: &sync.RWMutex{},
		},
		cronActions: actionCache{
			c: make(map[string][]Action),
			m: &sync.RWMutex{},
		},
		measurementActions: actionCache{
			c: make(map[string][]Action),
			m: &sync.RWMutex{},
		},
		apps: appCache{
			c: map[string]bool{},
			m: &sync.RWMutex{},
		},
	}
	c.populateTriggerCaches()

	return c
}

func (c *c) populateTriggerCaches() {

	r, err := c.db.Context().Query(`
		SELECT id
		FROM Application
	`)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Unable to list application.")
		return
	}
	defer r.Close()
	for r.Next() {
		var appId string
		err = r.Scan(&appId)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Unable to select application id.")
			return
		}
		log.Debug().Any("appId", appId).Msg("Found app.")
		c.apps.put(appId, true)
	}

	waiter := make(chan bool, 3)
	go func() {
		for _, appId := range c.apps.values() {
			tr := c.GetStatusChangeTrigger(appId)
			c.statusActions.put(appId, tr.Actions)
		}
		waiter <- true
	}()

	go func() {
		for _, appId := range c.apps.values() {
			mts := c.ListMeasurementTriggers(appId)
			for _, mt := range mts {
				c.measurementActions.put(fmt.Sprintf("%s:%s", appId, mt.MeasurementName), mt.Actions)
			}
		}
		waiter <- true
	}()

	<-waiter
	<-waiter
}

func startDatabase(dbLoc string) database.Database {
	return database.New(connString(dbLoc), ddl, "sqlite")
}

func connString(dbLoc string) string {
	return "file:///" + dbLoc + "?_pragma=foreign_keys(1)"
}

func (c *c) listActions(dba database.Context, anyTrigger any) ([]Action, error) {
	acts := make([]Action, 0)
	var column string
	var id int
	switch t := anyTrigger.(type) {
	case MeasurementTrigger:
		{
			column = "measurement_trigger_id"
			id = t.Id
		}
	case CronTrigger:
		{
			column = "cron_trigger_id"
			id = t.Id
		}
	case StatusChangeTrigger:
		{
			column = "status_change_trigger_id"
			id = t.Id
		}
	default:
		{
			err := fmt.Errorf("unable to handle action of type %s", t)
			log.Error().AnErr("error", err).Msg("Could not list actions.")
			return nil, err
		}
	}

	r, err := dba.Query(fmt.Sprintf(`
		SELECT
			name,
			script
		FROM Action
		WHERE %s = ?
		ORDER BY name ASC
	`, column), id)

	if err != nil {
		log.Error().AnErr("error", err).Msg("Could not list actions.")
		return nil, err
	}
	defer r.Close()
	for r.Next() {
		a := Action{}
		err = r.Scan(&a.Name, &a.Script)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list actions.")
			return nil, err
		}
		acts = append(acts, a)
	}
	return acts, nil
}

func (c *c) ListApplicationIds() []string {
	return c.apps.values()
}

func (c *c) UpdateStatusChangeTrigger(appId string, t StatusChangeTrigger) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
			DELETE FROM Action
			WHERE
				status_change_trigger_id = (
					SELECT id FROM
					StatusTrigger
					WHERE application_id = ?
				)`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete old actions.")
			return nil, err
		}

		var sId int

		r, err := dba.Query(`
			SELECT id
			FROM StatusTrigger
			WHERE application_id = ?
		`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not fetch trigger id.")
		}
		defer r.Close()
		if r.Next() {
			err = r.Scan(&sId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not fetch trigger id.")
			}
		}

		for _, a := range t.Actions {
			_, err := dba.Exec(`
				INSERT INTO Action (name, script, status_change_trigger_id)
				VALUES (?, ?, ?)	
			`, a.Name, a.Script, sId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}

		c.statusActions.put(appId, t.Actions)

		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list update trigger..")
	}
}

func (c *c) GetStatusChangeActions(appId string) []Action {
	return c.statusActions.get(appId)
}

func (c *c) GetStatusChangeTrigger(appId string) StatusChangeTrigger {
	tr, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		tr := StatusChangeTrigger{}
		r, err := dba.Query(`
			SELECT
				id
			FROM StatusTrigger
			WHERE application_id = ?`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list triggers.")
			return nil, err
		}
		defer r.Close()
		if r.Next() {
			err = r.Scan(&tr.Id)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("Database operation failed.")
			}
			a, err := c.listActions(dba, tr)
			if err != nil {
				return nil, err
			}
			tr.Actions = a
		}

		return tr, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list measurement triggers.")
	}

	return tr.(StatusChangeTrigger)
}

func (c *c) ListMeasurementTriggers(appId string) []MeasurementTrigger {
	ms, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		ms := make([]MeasurementTrigger, 0)
		r, err := dba.Query(`
			SELECT
				id,
				measurement_name
			FROM MeasurementTrigger
			WHERE application_id = ?
			`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list triggers.")
			return nil, err
		}
		defer r.Close()
		for r.Next() {
			m := MeasurementTrigger{}
			err = r.Scan(&m.Id, &m.MeasurementName)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("Database operation failed.")
			}
			a, err := c.listActions(dba, m)
			if err != nil {
				return nil, err
			}
			m.Actions = a
			ms = append(ms, m)
		}

		return ms, nil
	})

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list measurement triggers.")
	}

	return ms.([]MeasurementTrigger)
}

func (c *c) PersistStatusChangeTrigger(appId string, st StatusChangeTrigger) StatusChangeTrigger {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
			INSERT INTO StatusTrigger (application_id)
			VALUES (?)
		`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not insert status change trigger.")
			return nil, err
		}

		r, err := dba.Query(`
			SELECT id
			FROM StatusTrigger
			WHERE application_id = ?
		`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}
		defer r.Close()
		if !r.Next() {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}
		var stId int
		if err = r.Scan(&stId); err != nil {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}

		for _, a := range st.Actions {
			_, err := dba.Exec(`
				INSERT INTO Action (name, script, status_change_trigger_id)
				VALUES (?, ?, ?)
			`, a.Name, a.Script, stId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		c.statusActions.put(appId, st.Actions)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist status change trigger.")
	}

	return c.GetStatusChangeTrigger(appId)
}

func (c *c) PersitMeasurementTrigger(appId string, mt MeasurementTrigger) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
			INSERT INTO MeasurementTrigger (application_id, measurement_name)
			VALUES (?, ?)
		`, appId, mt.MeasurementName)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not insert measurement trigger.")
			return nil, err
		}
		r, err := dba.Query(`
			SELECT id
			FROM MeasurementTrigger
			WHERE application_id = ?
			AND measurement_name = ?
		`, appId, mt.MeasurementName)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Measurement trigger was not inserted.")
			return nil, err
		}
		defer r.Close()
		if !r.Next() {
			log.Error().AnErr("error", err).Msg("Measurement trigger was not inserted.")
			return nil, err
		}
		var mtId int
		if err = r.Scan(&mtId); err != nil {
			log.Error().AnErr("error", err).Msg("Measurement trigger was not inserted.")
			return nil, err
		}
		for _, a := range mt.Actions {
			_, err := dba.Exec(`
				INSERT INTO Action (name, script, measurement_trigger_id)
				VALUES (?, ?, ?)
			`, a.Name, a.Script, mtId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist measurement trigger.")
	}
}

// Public functions in alphabetical order
func (c *c) DeleteApplication(id string) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec("DELETE FROM Configuration WHERE application_id = ?", id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete configurations.")
			return nil, err
		}

		_, err = dba.Exec("DELETE FROM Application WHERE id = ?", id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete application.")
			return nil, err
		}
		c.apps.remove(id)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Deleting application failed.")
	}
}

func (c *c) GetApplication(id string) Application {
	a, err := c.getApplication(id, false)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get application")
	}
	return a
}

func (c *c) GetConfiguration(appId string) Configuration {
	conf := Configuration{}
	r, err := c.db.Context().Query(
		`SELECT 
			tmp.data,
			tmp.created
		FROM (
			SELECT 
				data,
				created 
			FROM 
				Configuration
				WHERE application_id = ?
				AND is_latest = true
			ORDER BY created DESC
			LIMIT 1
		) tmp
		`, appId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get configuration.")
	}
	defer r.Close()
	if r.Next() {
		err = r.Scan(&conf.Data, &conf.CreatedAt)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Could not get configuration.")
		}
	}
	return conf
}

func (c *c) IsAllowed(id string, apiKey string) bool {
	r, err := c.db.Context().Query("SELECT 1 FROM Application WHERE id = ? AND api_key = ?", id, apiKey)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func (c *c) ListApplications() []Application {
	result := make([]Application, 0)
	rows, err := c.db.Context().Query(
		`SELECT 
			id, 
			name,
			c.data,
			c.created
		FROM Application a
		LEFT JOIN (
			SELECT 
				data,
				created,
				application_id,
				is_latest,
				RANK() OVER (PARTITION BY application_id ORDER BY created DESC) latest_by_created
			FROM Configuration
		) AS c ON c.application_id = a.id AND c.latest_by_created = 1 AND c.is_latest = 1
		ORDER BY a.name`)

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Database operation failed.")
	}
	defer rows.Close()
	for rows.Next() {
		a := Application{
			Configuration: Configuration{},
		}
		err := rows.Scan(&a.Id, &a.Name, &a.Configuration.Data, &a.Configuration.CreatedAt)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Database operation failed.")
		}
		result = append(result, a)
	}
	return result
}

func (c *c) Login(login string, password string) bool {
	r, err := c.db.Context().Query("SELECT 1 FROM User WHERE login = ? AND password = ?", login, password)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func (c *c) PersistApplication(a Application) Application {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(
			`INSERT INTO Application
				(id, name, api_key)
			VALUES
				(?, ?, ?)
			`, a.Id, a.Name, a.ApiKey)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return nil, err
		} else {
			err = persistConfiguration(dba, a.Id, a.Configuration)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not persist configuration.")
				return nil, err
			}
		}
		c.apps.put(a.Id, true)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist application.")
	}

	a, err = c.getApplication(a.Id, true)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get application.")
	}
	return a
}

func (c *c) PersistConfiguration(applicationId string, conf Configuration) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		if err := persistConfiguration(dba, applicationId, conf); err != nil {
			log.Error().AnErr("error", err).Msg("Could not persist configuration.")
			return nil, err
		} else {
			return nil, nil
		}
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Persisting configuration failed.")
	}
}

func (c *c) UpdateApplication(a Application) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(
			`UPDATE Application
				SET
					name = ?
				WHERE
					id = ?`, a.Name, a.Id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return nil, err
		}
		if a.Configuration.Data != "" {
			err = persistConfiguration(dba, a.Id, a.Configuration)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not persist configuration.")
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not update application.")
	}
}

// Private functions in alphabetical order

func (c *c) getApplication(id string, withApiKey bool) (Application, error) {
	a := Application{
		Configuration: Configuration{},
	}
	r, err := c.db.Context().Query(
		`SELECT 
			a.id, 
			a.name,
			CASE
				WHEN ? THEN a.api_key
				ELSE ''
			END as api_key,
			c.data,
			c.created
		FROM Application a
		LEFT JOIN Configuration c ON c.application_id = a.id AND c.is_latest = true
		WHERE a.id = ?
		ORDER BY c.created DESC
		LIMIT 1
		`, withApiKey, id)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return a, err
	}
	defer r.Close()
	if r.Next() {
		err := r.Scan(&a.Id, &a.Name, &a.ApiKey, &a.Configuration.Data, &a.Configuration.CreatedAt)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return a, err
		}
	}
	return a, nil
}

func persistConfiguration(dba database.Context, applicationId string, c Configuration) error {
	_, err := dba.Exec(
		`UPDATE Configuration
			SET
				is_latest = false
			WHERE
				application_id = ?
				AND is_latest = true`, applicationId)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return err
	}
	_, err = dba.Exec(
		`INSERT INTO Configuration
			(application_id, data, is_latest)
		VALUES
			(?, ?, true)
		`, applicationId, c.Data)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return err
	}
	return nil
}

func (c *c) DeleteStatusChangeTrigger(appId string) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
			DELETE FROM Action
			WHERE status_change_trigger_id = (
				SELECT id
				FROM StatusTrigger
				WHERE application_id = ?
			)
		`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting actions failed.")
			return nil, err
		}
		_, err = dba.Exec(`
			DELETE FROM StatusTrigger
			WHERE application_id = ?
		`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting status change trigger failed.")
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Status change trigger deletion failed.")
	}
}
