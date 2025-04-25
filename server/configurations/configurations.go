package configurations

import (
	_ "embed"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

var db database.Database

// Public functions in alphabetical order
func DeleteApplication(id string) {

	_, err := db.DoInTransaction(func(dba database.Context) (any, error) {
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
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Deleting application failed.")
	}
}

func GetApplication(id string) Application {
	a, err := getApplication(id, false)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get application")
	}
	return a
}

func GetConfiguration(appId string) Configuration {
	c := Configuration{}
	r, err := db.Context().Query(
		`SELECT 
			data,
			created
		FROM Configuration
		WHERE application_id = ? AND is_latest = true
		`, appId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get configuration.")
	}
	defer r.Close()
	if r.Next() {
		err = r.Scan(&c.Data, &c.CreatedAt)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Could not get configuration.")
		}
	}
	return c
}

func IsAllowed(id string, apiKeyHash string) bool {
	r, err := db.Context().Query("SELECT 1 FROM Application WHERE id = ? AND api_key = ?", id, apiKeyHash)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func ListApplications() []Application {
	result := make([]Application, 0)
	rows, err := db.Context().Query(
		`SELECT 
			a.id, 
			a.name,
			c.data,
			c.created
		FROM Application a
		LEFT JOIN Configuration c ON c.application_id = a.id AND c.is_latest = true
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

func Login(login string, password string) bool {
	r, err := db.Context().Query("SELECT 1 FROM User WHERE login = ? AND password = ?", login, password)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func PersistApplication(a Application) Application {
	_, err := db.DoInTransaction(func(dba database.Context) (any, error) {
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
			log.Error().AnErr("error", err).Msg("Could not persist configuration.")
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist application.")
	}

	a, err = getApplication(a.Id, true)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not get application.")
	}
	return a
}

func PersistConfiguration(applicationId string, c Configuration) {
	_, err := db.DoInTransaction(func(dba database.Context) (any, error) {
		if err := persistConfiguration(dba, applicationId, c); err != nil {
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

func StartDatabase(dbLoc string) {
	if db != nil {
		log.Error().Msg("Database already started. You're good to go, but you might want to check what's going on.")
		return
	}
	db = database.New(connString(dbLoc), ddl, "sqlite")
}

func UpdateApplication(a Application) {
	_, err := db.DoInTransaction(func(dba database.Context) (any, error) {
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
func connString(dbLoc string) string {
	return "file:///" + dbLoc + "?_pragma=foreign_keys(1)"
}

func getApplication(id string, withApiKey bool) (Application, error) {
	a := Application{
		Configuration: Configuration{},
	}
	r, err := db.Context().Query(
		`SELECT 
			a.id, 
			a.name,
			CASE
				WHEN ? THEN a.api_key
				ELSE null
			END as api_key,
			c.data,
			c.created
		FROM Application a
		LEFT JOIN Configuration c ON c.application_id = a.id AND c.is_latest = true
		WHERE a.id = ?
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
