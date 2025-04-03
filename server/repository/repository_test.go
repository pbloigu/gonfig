package repository

import (
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Remove("/tmp/hello.db")
	StartDatabase("/tmp/hello.db")
	code := m.Run()
	os.Exit(code)
}

func TestRollback(t *testing.T) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Unable to start a transaction.")
		assert.Error(t, err)
	}
	tx.ExecContext(ctx, "INSERT INTO Application(id, name, api_key) values('test666','somename','keyy')")
	_, err = tx.ExecContext(ctx, "INSERT INTO Configuration(application_id) values ('hello')")
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	cnt := 99
	r, _ := db.Query("SELECT COUNT(*) FROM Application")
	if r.Next() {
		r.Scan(&cnt)
	} else {
		assert.Fail(t, "FAILURE")
	}

	assert.Equal(t, 0, cnt)
}

func TestPersistApplicationShouldRollback(t *testing.T) {
	defer func() {
		recover()
		r, _ := db.Query("SELECT COUNT (*) FROM Application")
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
