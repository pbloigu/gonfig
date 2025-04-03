package repository

import (
	"context"
	"database/sql"

	_ "embed"

	"github.com/rs/zerolog/log"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

var db *sql.DB

type dbWrapper struct {
	db  *sql.DB
	ctx context.Context
}

type txWrapper struct {
	tx  *sql.Tx
	ctx context.Context
}

type dbAccess interface {
	exec(query string, args ...any) (sql.Result, error)
	query(query string, args ...any) (*sql.Rows, error)
}

func (dbw dbWrapper) exec(query string, args ...any) (sql.Result, error) {
	return dbw.db.ExecContext(dbw.ctx, query, args...)
}

func (dbw dbWrapper) query(query string, args ...any) (*sql.Rows, error) {
	return dbw.db.QueryContext(dbw.ctx, query, args...)
}

func (txw txWrapper) exec(query string, args ...any) (sql.Result, error) {
	return txw.tx.ExecContext(txw.ctx, query, args...)
}

func (txw txWrapper) query(query string, args ...any) (*sql.Rows, error) {
	return txw.tx.QueryContext(txw.ctx, query, args...)
}

func wrap() dbWrapper {
	return dbWrapper{
		db:  db,
		ctx: context.Background(),
	}
}

// Public functions in alphabetical order
func DeleteApplication(id string) {

	_, err := doInTransaction(func(tba dbAccess) (any, error) {
		_, err := tba.exec("DELETE FROM Configuration WHERE application_id = ?", id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete configurations.")
			return nil, err
		}
		_, err = tba.exec("DELETE FROM Measurement WHERE application_id = ?", id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete measurements.")
			return nil, err
		}
		_, err = tba.exec("DELETE FROM Application WHERE id = ?", id)
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
	r, err := wrap().query(
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
	r, err := wrap().query("SELECT 1 FROM Application WHERE id = ? AND api_key = ?", id, apiKeyHash)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func ListApplications() []Application {
	result := make([]Application, 0)
	rows, err := wrap().query(
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

func GetMeasurement(applicationId string, measurementName string) Measurement {
	r, err := wrap().query(
		`SELECT
				m.id,
				m.name,
				(SELECT mv.created FROM MeasurementValue mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1),
				(SELECT mv.data FROM MeasurementValue mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1)
			FROM Measurement m		
			WHERE m.application_id = ?
			AND m.name = ?
	`, applicationId, measurementName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	if r.Next() {
		m := Measurement{}
		err := r.Scan(&m.Id, &m.Name, &m.LastValueTime, &m.LastValue)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		return m
	} else {
		return Measurement{}
	}
}

func ListMeasurementValues(measurementId int) []MeasurementValue {
	result := make([]MeasurementValue, 0)
	r, err := wrap().query(
		`SELECT
				created,
				data
			FROM MeasurementValue
			WHERE id = ?
			ORDER BY created DESC
		`, measurementId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	for r.Next() {
		m := MeasurementValue{}
		err := r.Scan(&m.CreatedAt, &m.Data)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		result = append(result, m)
	}
	return result
}

func ListMeasurements(applicationId string) []Measurement {
	result := make([]Measurement, 0)
	r, err := wrap().query(
		`SELECT
				m.id,
				m.name,
				(SELECT mv.created FROM MeasurementValue mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1),
				(SELECT mv.data FROM MeasurementValue mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1)
			FROM Measurement m		
			WHERE m.application_id = ?
			ORDER BY m.name ASC	
			`, applicationId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	for r.Next() {
		m := Measurement{}
		err := r.Scan(&m.Id, &m.Name, &m.LastValueTime, &m.LastValue)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		result = append(result, m)
	}
	return result
}

func Login(login string, password string) bool {
	r, err := wrap().query("SELECT 1 FROM User WHERE login = ? AND password = ?", login, password)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	return r.Next()
}

func PersistApplication(a Application) Application {
	_, err := doInTransaction(func(dba dbAccess) (any, error) {
		_, err := dba.exec(
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
	_, err := doInTransaction(func(dba dbAccess) (any, error) {
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

func PeristMeasurement(applicationId string, m Measurement) {
	_, err := doInTransaction(func(dba dbAccess) (any, error) {
		existing := GetMeasurement(applicationId, m.Name)
		if existing == (Measurement{}) {
			_, err := dba.exec(`
				INSERT INTO Measurement (application_id, name) VALUES (? ,?)
			`, applicationId, m.Name)
			if err != nil {
				log.Error().AnErr("error", err).Msg("SQL execution failed.")
				return nil, err
			}
			existing = GetMeasurement(applicationId, m.Name)
		}
		_, err := dba.exec(`
			INSERT INTO MeasurementValue (measurement_id, data) VALUES (?, ?)
		`, existing.Id, m.LastValue)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Persisting measurement value failed.")
	}

}

func StartDatabase(dbLoc string) {
	if db != nil {
		log.Error().Msg("Database already started. You're good to go, but you might want to check what's going on.")
		return
	}
	ctx := context.Background()
	o, err := sql.Open("sqlite", connString(dbLoc))

	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Failed to open database. This is unrecoverable.")
	} else {
		dbLogger := zerologadapter.New(log.Logger)
		db = sqldblogger.OpenDriver(connString(dbLoc), o.Driver(), dbLogger /*, using_default_options*/)
	}
	_, err = o.ExecContext(ctx, ddl)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Failed to initialize database. This is unrecoverable.")
	}
	log.Info().Any("location", dbLoc).Msg("Database online.")
}

func UpdateApplication(a Application) {
	_, err := doInTransaction(func(dba dbAccess) (any, error) {
		_, err := dba.exec(
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

func doInTransaction(f func(dba dbAccess) (any, error)) (any, error) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Unable to start a transaction.")
		return nil, err
	}
	defer tx.Rollback()
	txw := txWrapper{
		tx:  tx,
		ctx: ctx,
	}
	result, err := f(txw)
	if err != nil {
		return result, err
	}
	err = tx.Commit()
	if err != nil {
		log.Error().AnErr("error", err).Msg("Failed to commit database transaction.")
		return nil, err
	}
	return result, nil
}

func getApplication(id string, withApiKey bool) (Application, error) {
	a := Application{
		Configuration: Configuration{},
	}
	r, err := wrap().query(
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

func persistConfiguration(dba dbAccess, applicationId string, c Configuration) error {
	_, err := dba.exec(
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
	_, err = dba.exec(
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
