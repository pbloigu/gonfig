package database

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

type Database interface {
	DoInTransaction(f func(dba Context) (any, error)) (any, error)
	Context() Context
}

type Context interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

func New(uri string, ddl string, driver string) Database {

	ctx := context.Background()
	o, err := sql.Open(driver, uri)

	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Failed to open database. This is unrecoverable.")
	} else {
		dbLogger := zerologadapter.New(log.Logger)
		o = sqldblogger.OpenDriver(uri, o.Driver(), dbLogger /*, using_default_options*/)
	}
	_, err = o.ExecContext(ctx, ddl)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("Failed to initialize database. This is unrecoverable.")
	}
	log.Info().Any("location", uri).Msg("Database online.")
	return &database{
		db: o,
	}
}

type database struct {
	db *sql.DB
}

func (db database) Context() Context {
	return dbCtx{
		db:  db.db,
		ctx: context.Background(),
	}
}

type dbCtx struct {
	db  *sql.DB
	ctx context.Context
}

func (db dbCtx) Exec(query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(db.ctx, query, args...)
}

func (db dbCtx) Query(query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(db.ctx, query, args...)
}

type txCtx struct {
	tx  *sql.Tx
	ctx context.Context
}

func (tx txCtx) Exec(query string, args ...any) (sql.Result, error) {
	return tx.tx.ExecContext(tx.ctx, query, args...)
}

func (tx txCtx) Query(query string, args ...any) (*sql.Rows, error) {
	return tx.tx.QueryContext(tx.ctx, query, args...)
}

func (db database) DoInTransaction(f func(dba Context) (any, error)) (any, error) {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Unable to start a transaction.")
		return nil, err
	}
	defer tx.Rollback()
	txw := txCtx{
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
