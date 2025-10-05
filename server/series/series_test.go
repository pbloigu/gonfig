package series

import (
	"context"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

var container *mariadb.MariaDBContainer
var repo m

func TestMain(m *testing.M) {
	beforeAll()
	code := run(m)
	afterAll()
	os.Exit(code)
}

func run(m *testing.M) int {
	beforeEach()
	return m.Run()
}

func beforeEach() {
	clear()
}

func beforeAll() {
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
	container = tc
	repo = m{
		db: startDatabase(cstr),
	}
}

func afterAll() {
	if err := testcontainers.TerminateContainer(container); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}
}

func clear() {
	repo.db.Context().Exec("SET FOREIGN_KEY_CHECKS = 0")

	r, _ := repo.db.Context().Query(`
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
		repo.db.Context().Exec(s)
	}
	repo.db.Context().Exec("DELETE FROM Series")

}

func TestInitPersistDelete(t *testing.T) {
	repo.InitSeries("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	repo.PeristSeries("TestApp1", Series{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	m := repo.GetSeries("TestApp1", "testMeasurement1")
	assert.Equal(t, "testMeasurement1", m.Name)
	assert.Equal(t, testValue, *m.LastValue)
	repo.DeleteApplication("TestApp1")
}

func TestPersistMultiple(t *testing.T) {
	repo.InitSeries("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	repo.PeristSeries("TestApp1", Series{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	testValue = "testValue2"
	repo.PeristSeries("TestApp1", Series{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})

	m := repo.GetSeries("TestApp1", "testMeasurement1")
	values := repo.ListSeriesValues(m.Id, "created", "DESC", 1, 100)
	assert.Equal(t, 2, len(values))

}
