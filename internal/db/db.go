package db

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/pkg/errors"
)

type (
	// IDatabase represents the database client.
	IDatabase interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		Conn() *sql.DB
	}

	database struct {
		conn   *sql.DB
		logger slog.INamedLogger
	}
)

// New returns a new implementation of `IDatabase`.
func New(l slog.ILogger, location string) (IDatabase, error) {
	nl := slog.NewNamed(l, "diagnostics-db", slog.None)

	nl.Log("init", "msg", "creating .run directory if it does not exist")

	// -> Create the '.run' directory if it does not exist.
	dir := location + "/.run"
	src := dir + "/diagnostics.db"
	if err := fs.CreateDirINE(dir); err != nil {
		return nil, errors.Wrap(err, "diagnostics: failed to create '.run' directory")
	}

	nl.Log("open", "msg", "opening database connection", "src", src)

	// -> Open the database connection.
	conn, err := sql.Open("sqlite3", src)
	if err != nil {
		return nil, errors.Wrap(err, "diagnostics: failed to open database")
	}
	db := &database{
		conn:   conn,
		logger: nl,
	}

	// -> Seed the database.
	if err = db.seed(); err != nil {
		return nil, err
	}
	defer nl.Log("ready", "msg", "database ready")

	return db, nil
}

func (c *database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

func (c *database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (c *database) Conn() *sql.DB {
	return c.conn
}

func (c *database) prepare(query string) (*sql.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer c.logger.Log("exec", "query", query)

	return stmt, nil
}

func (c *database) seed() (err error) {
	c.logger.Log("seed", "msg", "seeding database")

	qs := []string{
		`
CREATE TABLE IF NOT EXISTS runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_name TEXT NOT NULL,
    models_checksum BLOB NOT NULL,
    interface_checksum BLOB NOT NULL
);`,
		`CREATE INDEX IF NOT EXISTS file_name ON runs (file_name);`,
	}

	for i, q := range qs {
		if _, err = c.ExecContext(context.Background(), q); err != nil {
			err = errors.Wrapf(err, "diagnostics: failed to seed database: stmt[%d]", i)
		}
	}
	return
}
