package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	goose "github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

type zerologBridge struct {
}

func (zlb zerologBridge) Fatalf(format string, v ...interface{}) {
	log.Fatal().Msg(fmt.Sprintf(format, v))
}

func (zlb zerologBridge) Printf(format string, v ...interface{}) {
	log.Info().Msg(fmt.Sprintf(format, v))
}

var mgrDialects = map[string]goose.Dialect{
	"mysql":  goose.DialectMySQL,
	"sqlite": goose.DialectSQLite3,
}

type Database interface {
	DoInTransaction(f func(dba Context) (any, error)) (any, error)
	Context() Context
	Db() *sql.DB
}

type Context interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

func New(uri string, ddl fs.FS, driver string) Database {

	o, err := sql.Open(driver, uri)

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Failed to open database. This is unrecoverable.")
	} else {
		dbLogger := zerologadapter.New(log.Logger)
		o = sqldblogger.OpenDriver(uri, o.Driver(), dbLogger /*, using_default_options*/)
	}
	if err = migrate(o, ddl, mgrDialects[driver]); err != nil {
		log.Panic().AnErr("error", err).Msg("Failed to apply database migrations.")
	}
	log.Info().Any("location", uri).Msg("Database online.")
	return &database{
		db: o,
	}
}

func migrate(db *sql.DB, ddl fs.FS, dialect goose.Dialect) error {

	p, err := goose.NewProvider(dialect, db, ddl, goose.WithVerbose(true), goose.WithLogger(&zerologBridge{}))
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = p.Up(ctx)
	if err != nil {
		return err
	} else {
		log.Info().Msg("Database was migrated.")
		return nil
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

func (db database) Db() *sql.DB {
	return db.db
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
