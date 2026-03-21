package configurations

import (
	"embed"
	_ "embed"
	"io/fs"
	"sync"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

type Configurations interface {
	// Application related
	PersistApplication(a Application) Application
	DeleteApplication(id string)
	GetApplication(id string) Application
	ListApplications() []Application
	UpdateApplication(a Application)
	ListApplicationIds() []string

	// Configuration related
	PersistConfiguration(applicationId string, c Configuration)
	GetConfiguration(appId string) Configuration

	// Access management
	IsAllowed(id string, apiKey string) bool
	Login(login string, password string) bool

	// Series triggers
	ListSeriesTriggers(appId string) []SeriesTrigger
	PersistSeriesTrigger(appId string, mt SeriesTrigger)

	// Status change triggers
	PersistStatusChangeTrigger(appId string, st StatusChangeTrigger) StatusChangeTrigger
	UpdateStatusChangeTrigger(appId string, t StatusChangeTrigger)
	GetStatusChangeTrigger(appId string) StatusChangeTrigger
	DeleteStatusChangeTrigger(appId string)
	GetStatusChangeActions(appId string) []Action

	// Cron triggers
	PersistCronTrigger(cr CronTrigger) CronTrigger
	GetCronTrigger(id int) CronTrigger
	DeleteCronTrigger(id int)
	UpdateCronTrigger(cr CronTrigger)
	ListCronTriggers() []CronTrigger
}

type c struct {
	db             database.Database
	statusTriggers *sync.Map
	seriesTriggers *sync.Map
	apps           *sync.Map
}

//go:embed db/*.sql
var ddls embed.FS

// New constructs a Configurations backed by the database at dbLoc.
func New(dbLoc string) Configurations {
	c := &c{
		db:             startDatabase(dbLoc),
		statusTriggers: &sync.Map{},
		seriesTriggers: &sync.Map{},
		apps:           &sync.Map{},
	}
	c.populateTriggerCaches()

	return c
}

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
		c.apps.Delete(id)
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
            data,
            created 
        FROM 
            Configuration
        WHERE application_id = ?
        AND is_latest = true
        ORDER BY created DESC
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
		c.apps.Store(a.Id, true)
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

func (c *c) ListApplicationIds() []string {
	ret := make([]string, 0)
	c.apps.Range(func(key, value any) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func startDatabase(dbLoc string) database.Database {
	sub, err := fs.Sub(ddls, "db")
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Failed to access migration files.")
	}
	return database.New(connString(dbLoc), sub, "sqlite")
}

func connString(dbLoc string) string {
	return "file:///" + dbLoc + "?_pragma=foreign_keys(1)"
}

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
