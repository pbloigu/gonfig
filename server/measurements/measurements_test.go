package measurements

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	godotenv.Load(".env.test")
	StartDatabase(os.Getenv("GONFIG_MEASUREMENT_DB"))
	// clear()
}

func clear() {
	db.Context().Exec("SET FOREIGN_KEY_CHECKS = 0")

	r, _ := db.Context().Query(`
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
		db.Context().Exec(s)
	}
	db.Context().Exec("DELETE FROM Measurement")

}

func TestInitPersistDelete(t *testing.T) {
	InitMeasurement("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	m := GetMeasurement("TestApp1", "testMeasurement1")
	assert.Equal(t, "testMeasurement1", m.Name)
	assert.Equal(t, testValue, *m.LastValue)
	DeleteApplication("TestApp1")
}

func TestPersistMultiple(t *testing.T) {
	InitMeasurement("TestApp1", "testMeasurement1")
	testValue := "testValue1"
	PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})
	testValue = "testValue2"
	PeristMeasurement("TestApp1", Measurement{
		Name:      "testMeasurement1",
		LastValue: &testValue,
	})

	m := GetMeasurement("TestApp1", "testMeasurement1")
	values := ListMeasurementValues(m.Id, "created", "DESC", 1, 100)
	assert.Equal(t, 2, len(values))

}

func TestGenerateSomeData(t *testing.T) {
	var val string
	for i := range 20 {
		val = fmt.Sprintf("value_%d", i)
		PeristMeasurement("ddc22738-b45e-4922-81f8-77017e800eb8", Measurement{
			Name:      "test1",
			LastValue: &val,
		})
	}
}
