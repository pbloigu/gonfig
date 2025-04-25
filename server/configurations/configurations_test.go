package configurations

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Remove("/tmp/hello.db")
	StartDatabase("/tmp/hello.db")
	code := m.Run()
	os.Exit(code)
}

func TestPersistApplicationShouldRollback(t *testing.T) {
	defer func() {
		recover()
		r, _ := db.Context().Query("SELECT COUNT (*) FROM Application")
		count := 1
		r.Next()
		r.Scan(&count)
		assert.Equal(t, 0, count)
	}()

	PersistApplication(Application{
		Id:     "hello123",
		Name:   "fooman",
		ApiKey: "apikey",
		Configuration: Configuration{
			Data: "",
		},
	})
}
