package diagnostics

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/pkg/errors"
)

type (
	// IDatabase represents the database client.
	IDatabase interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		Conn() *sql.DB
	}

	databaseClient struct {
		conn   *sql.DB
		logger slog.INamedLogger
	}
)

// NewDB returns a new implementation of `IDatabase`.
func newDB(l slog.ILogger, location string) (IDatabase, error) {
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
	db, err := sql.Open("sqlite3", src)
	if err != nil {
		return nil, errors.Wrap(err, "diagnostics: failed to open database")
	}
	c := &databaseClient{
		conn:   db,
		logger: nl,
	}

	// -> Seed the database.
	if err = c.seed(); err != nil {
		return nil, err
	}
	defer nl.Log("ready", "msg", "database ready")

	return c, nil
}

func (c *databaseClient) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

func (c *databaseClient) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (c *databaseClient) Conn() *sql.DB {
	return c.conn
}

func (c *databaseClient) prepare(query string) (*sql.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer c.logger.Log("exec", "query", query)

	return stmt, nil
}

func (c *databaseClient) seed() (err error) {
	c.logger.Log("seed", "msg", "seeding database")

	q := `
CREATE TABLE IF NOT EXISTS runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    models_checksum BLOB NOT NULL,
    interface_checksum BLOB NOT NULL
);`
	if _, err = c.ExecContext(context.Background(), q); err != nil {
		err = errors.Wrap(err, "diagnostics: failed to seed database")
	}
	return
}
