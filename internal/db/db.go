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
		RequiresSeed() bool
		Conn() *sql.DB
	}

	database struct {
		conn   *sql.DB
		logger slog.INamedLogger
		seed   bool
	}
)

// New returns a new implementation of `IDatabase`.
func New(l slog.ILogger, location string) (IDatabase, error) {
	nl := slog.NewNamed(l, "diagnostics-db", slog.None)

	nl.Log("init", "msg", "creating .run directory if it does not exist")

	// -> Create the '.run' directory if it does not exist.
	dir := location + "/.run"
	src := dir + "/diagnostics.db"
	ok, err := fs.CreateDirINE(dir)
	if err != nil {
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
		seed:   ok,
	}
	defer nl.Log("ready", "msg", "database ready")

	return db, nil
}

func (c *database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

func (c *database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := c.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (c *database) Conn() *sql.DB {
	return c.conn
}

func (c *database) RequiresSeed() bool {
	return c.seed
}
