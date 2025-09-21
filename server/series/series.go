package series

import (
	_ "embed"
	"fmt"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var ddl string

type SeriesDb interface {
	GetSeries(applicationId string, seriesName string) Series
	CountSeriesValues(seriesId int) int
	ListSeriesValues(seriesId int, sort string, dir string, page int, pageSize int) []SeriesValue
	ListSeries(applicationId string) []Series
	InitSeries(applicationId string, seriesName string)
	PeristSeries(applicationId string, m Series)
	DeleteApplication(id string)
}

type m struct {
	db database.Database
}

func New(connectionString string) SeriesDb {
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

func getSeries(dba database.Context, applicationId string, seriesName string) (Series, error) {

	r, err := dba.Query(`SELECT id FROM Series WHERE application_id = ? AND name = ?`, applicationId, seriesName)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return Series{}, err
	}
	defer r.Close()
	if !r.Next() {
		err = fmt.Errorf("no series found with the name %s for application %s", seriesName, applicationId)
		log.Error().AnErr("error", err).Msg("No series found.")
		return Series{}, nil
	}
	var id int
	if err = r.Scan(&id); err != nil {
		log.Error().AnErr("error", err).Msg("No series found.")
		return Series{}, err
	}
	r.Close()

	r, err = dba.Query(fmt.Sprintf(
		`SELECT
				m.id,
				m.name,
				(SELECT mv.created FROM SeriesValue_%d mv WHERE mv.series_id = m.id ORDER BY mv.created DESC LIMIT 1),
				(SELECT mv.data FROM SeriesValue_%d mv WHERE mv.series_id = m.id ORDER BY mv.created DESC LIMIT 1)
			FROM Series m		
			WHERE m.application_id = ?
			AND m.name = ?
	`, id, id), applicationId, seriesName)
	if err != nil {
		log.Error().AnErr("error", err).Msg("SQL execution failed.")
		return Series{}, err
	}
	defer r.Close()
	if r.Next() {
		m := Series{}
		err := r.Scan(&m.Id, &m.Name, &m.LastValueTime, &m.LastValue)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return Series{}, err
		}
		return m, nil
	} else {
		return Series{}, nil
	}
}

func (m *m) GetSeries(applicationId string, seriesName string) Series {
	series, err := getSeries(m.db.Context(), applicationId, seriesName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Fetching series failed.")
		return Series{}
	}
	return series
}

func (m *m) CountSeriesValues(seriesId int) int {
	r, err := m.db.Context().Query(fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM SeriesValue_%d 
        WHERE series_id = ?`, seriesId), seriesId)
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

func (m *m) ListSeriesValues(seriesId int, sort string, dir string, page int, pageSize int) []SeriesValue {
	result := make([]SeriesValue, 0)
	r, err := m.db.Context().Query(fmt.Sprintf(
		`SELECT
                created,
                data
            FROM SeriesValue_%d
            WHERE series_id = ?
            ORDER BY %s %s
            LIMIT %d
            OFFSET %d
        `, seriesId, sort, dir, pageSize, (page-1)*pageSize), seriesId)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	for r.Next() {
		m := SeriesValue{}
		err := r.Scan(&m.CreatedAt, &m.Data)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		result = append(result, m)
	}
	return result
}

func (m *m) ListSeries(applicationId string) []Series {
	result := make([]Series, 0)

	r, err := m.db.Context().Query(`SELECT id FROM Series where application_id = ?`, applicationId)
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
                    (SELECT mv.created FROM SeriesValue_%d mv WHERE mv.series_id = m.id ORDER BY mv.created DESC LIMIT 1),
                    (SELECT mv.data FROM SeriesValue_%d mv WHERE mv.series_id = m.id ORDER BY mv.created DESC LIMIT 1)
                FROM Series m		
                WHERE m.application_id = ?
                ORDER BY m.name ASC	
                `, id, id), applicationId)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("SQL execution failed.")
		}
		defer r.Close()

		for r.Next() {

			m := Series{}
			err := r.Scan(&m.Id, &m.Name, &m.LastValueTime, &m.LastValue)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("SQL execution failed.")
			}
			result = append(result, m)
		}
	}
	return result
}

func (m *m) InitSeries(applicationId string, seriesName string) {
	_, err := m.db.Context().Exec(`
                INSERT INTO Series (application_id, name) VALUES (? ,?)
            `,
		applicationId, seriesName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	r, err := m.db.Context().Query(
		`SELECT id FROM Series WHERE application_id = ? AND name = ?
        `, applicationId, seriesName)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("SQL execution failed.")
	}
	defer r.Close()
	var id int
	if !r.Next() {
		log.Panic().AnErr("error", err).Msg("Newly created series not found.")
	}
	if err = r.Scan(&id); err != nil {
		log.Panic().AnErr("error", err).Msg("No id found for newly created series.")
	}

	_, err = m.db.Context().Exec(fmt.Sprintf(
		`CREATE TABLE SeriesValue_%d (
            id int NOT NULL AUTO_INCREMENT,
            series_id int NOT NULL,
            created INT(11) UNSIGNED default UNIX_TIMESTAMP(),
            recorded int(11) UNSIGNED,
            data varchar(128),
            PRIMARY KEY (id),
            FOREIGN KEY (series_id) REFERENCES Series(id))
        `, id))
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not create table for storing series values.")
	}

}

func (m *m) PeristSeries(applicationId string, series Series) {
	_, err := m.db.DoInTransaction(func(dba database.Context) (any, error) {
		existing, err := getSeries(dba, applicationId, series.Name)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Fetching series failed.")
			return nil, err
		}
		if existing == (Series{}) {
			err = fmt.Errorf("no series found with the name %s for application %s", series.Name, applicationId)
			log.Error().AnErr("error", err).Msg("No series found.")
			return nil, err
		}
		_, err = dba.Exec(fmt.Sprintf(`
            INSERT INTO SeriesValue_%d (series_id, data) VALUES (?, ?)
        `, existing.Id), existing.Id, series.LastValue)
		if err != nil {
			log.Error().AnErr("error", err).Msg("SQL execution failed.")
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Persisting series value failed.")
	}
}

func (m *m) DeleteApplication(id string) {
	_, err := m.db.DoInTransaction(func(dba database.Context) (any, error) {
		r, err := dba.Query(
			`SELECT id FROM Series
                WHERE application_id = ?			
            `, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not enumerate series.")
			return nil, err
		}
		defer r.Close()
		ids := make([]int, 0)
		for r.Next() {
			var id int
			err = r.Scan(&id)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not select series id to delete.")
				return nil, err
			}
			ids = append(ids, id)
		}

		for _, id := range ids {
			_, err = dba.Exec(fmt.Sprintf(`DROP TABLE SeriesValue_%d`, id))
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not drop value table.")
				return nil, err
			}
		}

		_, err = dba.Exec(
			`DELETE 
				FROM Series
				WHERE application_id = ?
			`, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete series.")
			return nil, err
		}
		return nil, nil

	})

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Deleting application failed.")
	}
}
