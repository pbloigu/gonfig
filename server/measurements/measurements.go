package measurements

import (
	_ "embed"
	"fmt"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var ddl string

type Measurements interface {
	GetMeasurement(applicationId string, measurementName string) Measurement
	CountMeasurementValues(measurementId int) int
	ListMeasurementValues(measurementId int, sort string, dir string, page int, pageSize int) []MeasurementValue
	ListMeasurements(applicationId string) []Measurement
	InitMeasurement(applicationId string, measurementName string)
	PeristMeasurement(applicationId string, m Measurement)
	DeleteApplication(id string)
}

type m struct {
	db database.Database
}

func New(connectionString string) Measurements {
	return &m{
		db: startDatabase(connectionString),
	}
}

func startDatabase(connectionString string) database.Database {
	return database.New(connString(connectionString), ddl, "mysql")
}

func connString(connectionString string) string {
	return connectionString + "?multiStatements=true&parseTime=true"
}

func getMeasurement(dba database.Context, applicationId string, measurementName string) (Measurement, error) {

	r, err := dba.Query(`SELECT id from Measurement WHERE application_id = ? AND name = ?`, applicationId, measurementName)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return Measurement{}, err
	}
	defer r.Close()
	if !r.Next() {
		err = fmt.Errorf("no measurement found with the name %s for application %s", measurementName, applicationId)
		log.Error().AnErr("error", err).Msg("No measurement found.")
		return Measurement{}, err
	}
	var id int
	if err = r.Scan(&id); err != nil {
		log.Error().AnErr("error", err).Msg("No measurement found.")
		return Measurement{}, err
	}
	r.Close()

	r, err = dba.Query(fmt.Sprintf(
		`SELECT
				m.id,
				m.name,
				(SELECT mv.created FROM MeasurementValue_%d mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1),
				(SELECT mv.data FROM MeasurementValue_%d mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1)
			FROM Measurement m		
			WHERE m.application_id = ?
			AND m.name = ?
	`, id, id), applicationId, measurementName)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return Measurement{}, err
	}
	defer r.Close()
	if r.Next() {
		m := Measurement{}
		err := r.Scan(&m.Id, &m.Name, &m.LastValueTime, &m.LastValue)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return Measurement{}, err
		}
		return m, nil
	} else {
		return Measurement{}, nil
	}
}

func (m *m) GetMeasurement(applicationId string, measurementName string) Measurement {
	measurement, err := getMeasurement(m.db.Context(), applicationId, measurementName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Fetching measurement failed.")
		return Measurement{}
	}
	return measurement
}

func (m *m) CountMeasurementValues(measurementId int) int {
	r, err := m.db.Context().Query(fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM MeasurementValue_%d 
        WHERE measurement_id = ?`, measurementId), measurementId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	r.Next()
	var cnt int
	if err = r.Scan(&cnt); err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	return cnt
}

func (m *m) ListMeasurementValues(measurementId int, sort string, dir string, page int, pageSize int) []MeasurementValue {
	result := make([]MeasurementValue, 0)
	r, err := m.db.Context().Query(fmt.Sprintf(
		`SELECT
                created,
                data
            FROM MeasurementValue_%d
            WHERE measurement_id = ?
            ORDER BY %s %s
            LIMIT %d
            OFFSET %d
        `, measurementId, sort, dir, pageSize, (page-1)*pageSize), measurementId)
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

func (m *m) ListMeasurements(applicationId string) []Measurement {
	result := make([]Measurement, 0)

	r, err := m.db.Context().Query(`SELECT id FROM Measurement where application_id = ?`, applicationId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	var id int
	for r.Next() {
		if err = r.Scan(&id); err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		r, err = m.db.Context().Query(fmt.Sprintf(
			`SELECT
                    m.id,
                    m.name,
                    (SELECT mv.created FROM MeasurementValue_%d mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1),
                    (SELECT mv.data FROM MeasurementValue_%d mv WHERE mv.measurement_id = m.id ORDER BY mv.created DESC LIMIT 1)
                FROM Measurement m		
                WHERE m.application_id = ?
                ORDER BY m.name ASC	
                `, id, id), applicationId)
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
	}
	return result
}

func (m *m) InitMeasurement(applicationId string, measurementName string) {
	_, err := m.db.Context().Exec(`
                INSERT INTO Measurement (application_id, name) VALUES (? ,?)
            `,
		applicationId, measurementName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	r, err := m.db.Context().Query(
		`SELECT id FROM Measurement WHERE application_id = ? AND name = ?
        `, applicationId, measurementName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	var id int
	if !r.Next() {
		log.Panic().AnErr("error", err).Msg("Newly created measurement not found.")
	}
	if err = r.Scan(&id); err != nil {
		log.Panic().AnErr("error", err).Msg("No id found for newly created measurement.")
	}

	_, err = m.db.Context().Exec(fmt.Sprintf(
		`CREATE TABLE MeasurementValue_%d (
            id int NOT NULL AUTO_INCREMENT,
            measurement_id int NOT NULL,
            created INT(11) UNSIGNED default UNIX_TIMESTAMP(),
            recorded int(11) UNSIGNED,
            data varchar(128),
            PRIMARY KEY (id),
            FOREIGN KEY (measurement_id) REFERENCES Measurement(id))
        `, id))
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not create table for storing measurement values.")
	}

}

func (m *m) PeristMeasurement(applicationId string, measurement Measurement) {
	_, err := m.db.DoInTransaction(func(dba database.Context) (any, error) {
		existing, err := getMeasurement(dba, applicationId, measurement.Name)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Fetching measurement failed.")
			return nil, err
		}
		if existing == (Measurement{}) {
			err = fmt.Errorf("no measurement found with the name %s for application %s", measurement.Name, applicationId)
			log.Error().AnErr("error", err).Msg("No measurement found.")
			return nil, err
		}
		_, err = dba.Exec(fmt.Sprintf(`
            INSERT INTO MeasurementValue_%d (measurement_id, data) VALUES (?, ?)
        `, existing.Id), existing.Id, measurement.LastValue)
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

func (m *m) DeleteApplication(id string) {
	_, err := m.db.DoInTransaction(func(dba database.Context) (any, error) {
		r, err := dba.Query(
			`SELECT id FROM Measurement
                WHERE application_id = ?			
            `, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not enumerate measurements.")
			return nil, err
		}
		defer r.Close()
		ids := make([]int, 0)
		for r.Next() {
			var id int
			err = r.Scan(&id)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not select measurement id to delete.")
				return nil, err
			}
			ids = append(ids, id)
		}

		for _, id := range ids {
			_, err = dba.Exec(fmt.Sprintf(`DROP TABLE MeasurementValue_%d`, id))
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not drop value table.")
				return nil, err
			}
		}

		_, err = dba.Exec(
			`DELETE 
				FROM Measurement
				WHERE application_id = ?
			`, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete measurements.")
			return nil, err
		}
		return nil, nil

	})

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Deleting application failed.")
	}
}
