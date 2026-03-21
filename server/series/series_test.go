package series

import (
	"context"
	"fmt"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

type SeriesTestSuite struct {
	suite.Suite
	container *mariadb.MariaDBContainer
	repo      m
}

func TestSeriesSuite(t *testing.T) {
	suite.Run(t, &SeriesTestSuite{})
}

func (st *SeriesTestSuite) SetupSuite() {
	ctx := context.Background()
	tc, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
	)
	if err != nil {
		panic(err)
	}
	cstr, err := tc.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	st.container = tc
	st.repo = m{
		db: startDatabase(cstr),
	}
}

func (st *SeriesTestSuite) TearDownSuite() {
	if err := testcontainers.TerminateContainer(st.container); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}
}

func (st *SeriesTestSuite) SetupTest() {
	st.repo.db.Context().Exec("SET FOREIGN_KEY_CHECKS = 0")

	r, _ := st.repo.db.Context().Query(`
		SELECT concat('DROP TABLE IF EXISTS ', table_name)
		FROM information_schema.tables
		WHERE table_schema = 'test'
		AND table_name like 'SeriesValue%'`)

	sqls := make([]string, 0)
	for r.Next() {
		var sql string
		r.Scan(&sql)
		sqls = append(sqls, sql)
	}
	r.Close()

	for _, s := range sqls {
		st.repo.db.Context().Exec(s)
	}
	st.repo.db.Context().Exec("DELETE FROM Series")
}

func (st *SeriesTestSuite) TestGetSeries() {
	seriesId := st.insertSeries("appId", "series1")
	var lastValue string
	for i := range 2 {
		lastValue = fmt.Sprintf("value: %d", i)
		st.insertValue(seriesId, i*100, lastValue)
	}

	series := st.repo.GetSeries("appId", "series1")
	st.NotEqual(0, series.Id)
	st.Equal("series1", series.Name)
	st.Equal(lastValue, *series.LastValue)

	series = st.repo.GetSeries("appId1", "seriesId2")
	st.Equal(0, series.Id)

	series = st.repo.GetSeries("appId2", "seriesId1")
	st.Equal(0, series.Id)
}

func (st *SeriesTestSuite) TestCountSeriesValues() {
	seriesId := st.insertSeries("appId", "series1")
	for i := range 51 {
		st.insertValue(seriesId, i*100, fmt.Sprintf("value: %d", i))
	}
	st.Equal(51, st.repo.CountSeriesValues(seriesId))
}

func (st *SeriesTestSuite) TestListSeriesValues() {
	seriesId := st.insertSeries("appId", "series1")
	values := make([]SeriesValue, 0)
	for i := range 51 {
		recdAt := i * 100
		data := fmt.Sprintf("value: %d", i)
		st.insertValue(seriesId, recdAt, data)
		values = append(values, SeriesValue{
			RecordedAt: &recdAt,
			Data:       &data,
		})
	}

	result := st.repo.ListSeriesValues(seriesId, Sort{Field: ID, Dir: DESC}, Pagination{Page: 1, Size: 10})
	st.Len(result, 10)
	st.Equal(51, result[0].Id)
}

func (st *SeriesTestSuite) TestListSeries() {
	seriesId := st.insertSeries("appId", "series1")
	for i := range 9 {
		st.insertValue(seriesId, i*100, fmt.Sprintf("value: %d", i))
	}

	seriesId = st.insertSeries("appId", "series2")
	for i := range 9 {
		st.insertValue(seriesId, i*100, fmt.Sprintf("value: %d", i))
	}

	series := st.repo.ListSeries("appId")
	st.Len(series, 2)
	st.Equal("series1", series[0].Name)
	st.Equal("value: 8", *series[0].LastValue)

	st.Equal("series2", series[1].Name)
	st.Equal("value: 8", *series[1].LastValue)
}

func (st *SeriesTestSuite) TestInitSeries() {
	st.repo.InitSeries("appId", "series1")
	r, err := st.repo.db.Context().Query(`SELECT id FROM Series where name = ? AND application_id = ?`, "series1", "appId")
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	st.True(r.Next())
	var id int
	r.Scan(&id)
	r.Close()

	r, err = st.repo.db.Db().Query(fmt.Sprintf(`SELECT COUNT(*) FROM SeriesValue_%d`, id))
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	st.True(r.Next())
	var count = -1
	r.Scan(&count)
	r.Close()
	st.Equal(0, count)
}

func (st *SeriesTestSuite) TestPeristSeriesValue() {
	seriesId := st.insertSeries("appId", "series1")
	data := "hello"
	st.repo.PeristSeriesValue("appId", "series1", SeriesValue{
		Data: &data,
	})

	r, err := st.repo.db.Db().Query(fmt.Sprintf(`SELECT data FROM SeriesValue_%d`, seriesId))
	if err != nil {
		st.Fail(err.Error())
	}
	defer r.Close()
	st.True(r.Next())
	r.Scan(&data)
	st.Equal("hello", data)
}

func (st *SeriesTestSuite) TestDeleteApplication() {
	seriesId := st.insertSeries("appId", "series1")
	st.insertValue(seriesId, 100, "somevalue")

	st.repo.DeleteApplication("appId")

	r, err := st.repo.db.Context().Query(`SELECT COUNT(*) FROM Series WHERE application_id = ?`, "appId")
	if err != nil {
		st.Fail(err.Error())
	}
	st.True(r.Next())
	cnt := -1
	r.Scan(&cnt)
	st.Equal(0, cnt)
	r.Close()

	r, err = st.repo.db.Context().Query(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'test'
		AND table_name like 'SeriesValue_%d'`, seriesId))
	if err != nil {
		st.Fail(err.Error())
	}
	st.True(r.Next())
	cnt = -1
	r.Scan(&cnt)
	st.Equal(0, cnt)
	r.Close()

}

func (st *SeriesTestSuite) TestInitPersistDelete() {
	st.repo.InitSeries("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	testRecorded := 1768723151
	st.repo.PeristSeriesValue("TestApp1", "testMeasurement1", SeriesValue{
		Data:       &testValue,
		RecordedAt: &testRecorded,
	})
	m := st.repo.GetSeries("TestApp1", "testMeasurement1")
	st.Equal("testMeasurement1", m.Name)
	st.Equal(testValue, *m.LastValue)
	st.Equal(testRecorded, *m.LastValueRecorded)
	st.repo.DeleteApplication("TestApp1")
}

func (st *SeriesTestSuite) TestPersistMultiple() {
	st.repo.InitSeries("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	testRecorded := 1768723151
	st.repo.PeristSeriesValue("TestApp1", "testMeasurement1", SeriesValue{
		Data:       &testValue,
		RecordedAt: &testRecorded,
	})
	testValue = "testValue2"
	testRecorded = 1768723152
	st.repo.PeristSeriesValue("TestApp1", "testMeasurement1", SeriesValue{
		Data:       &testValue,
		RecordedAt: &testRecorded,
	})

	m := st.repo.GetSeries("TestApp1", "testMeasurement1")
	values := st.repo.ListSeriesValues(m.Id, Sort{Field: CREATED, Dir: ASC}, Pagination{Page: 1, Size: 10})
	st.Len(values, 2)
	st.Equal("testValue1", *values[0].Data)
	st.Equal("testValue2", *values[1].Data)
	st.Equal(1768723151, *values[0].RecordedAt)
	st.Equal(1768723152, *values[1].RecordedAt)

}

func (st *SeriesTestSuite) insertSeries(appId string, seriesName string) int {
	st.repo.db.Context().Exec(`INSERT INTO Series (application_id, name) VALUES (?, ?)`, appId, seriesName)
	r, err := st.repo.db.Context().Query(`SELECT id FROM Series WHERE application_id = ? AND name = ?`, appId, seriesName)
	var seriesId int
	if err == nil && r.Next() {
		defer r.Close()
		r.Scan(&seriesId)
	} else {
		st.Fail(err.Error())
	}
	st.repo.db.Context().Exec(fmt.Sprintf(
		`CREATE TABLE SeriesValue_%d (
            id int NOT NULL AUTO_INCREMENT,
            series_id int NOT NULL,
            created INT(11) UNSIGNED default UNIX_TIMESTAMP(),
            recorded int(11) UNSIGNED,
            data varchar(128),
            PRIMARY KEY (id),
            FOREIGN KEY (series_id) REFERENCES Series(id))
        `, seriesId))
	return seriesId
}

func (st *SeriesTestSuite) insertValue(seriesId int, recorded int, value string) {
	st.repo.db.Context().Exec(fmt.Sprintf(`
		INSERT INTO SeriesValue_%d (series_id, recorded, data) VALUES (?, ?, ?) 
	`, seriesId), seriesId, recorded, value)
}
