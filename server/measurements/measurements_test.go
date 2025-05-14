package measurements

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
		AND table_name like 'MeasurementValue%'`)

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
	repo.db.Context().Exec("DELETE FROM Measurement")

}

func TestInitPersistDelete(t *testing.T) {
	repo.InitMeasurement("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	repo.PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	m := repo.GetMeasurement("TestApp1", "testMeasurement1")
	assert.Equal(t, "testMeasurement1", m.Name)
	assert.Equal(t, testValue, *m.LastValue)
	repo.DeleteApplication("TestApp1")
}

func TestPersistMultiple(t *testing.T) {
	repo.InitMeasurement("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	repo.PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	testValue = "testValue2"
	repo.PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})

	m := repo.GetMeasurement("TestApp1", "testMeasurement1")
	values := repo.ListMeasurementValues(m.Id, "created", "DESC", 1, 100)
	assert.Equal(t, 2, len(values))

}

// func TestGenerateSomeData(t *testing.T) {
// 	var val string
// 	for i := range 20 {
// 		val = fmt.Sprintf("value_%d", i)
// 		PeristMeasurement("3db49a19-1f9e-4a63-a374-16567a3e1ce4", Measurement{
// 			Name:      "test1",
// 			LastValue: &val,
// 		})
// 	}
// }
