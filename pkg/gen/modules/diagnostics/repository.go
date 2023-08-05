package diagnostics

import (
	"context"
	"database/sql"
	"github.com/maxzaleski/codegen/internal/db"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/pkg/errors"
)

type (
	IRepository interface {
		SeedDB() error
		FindOne(ctx context.Context, pkg string, pi int) (*snapshot, error)
		InsertOne(ctx context.Context, pkg string, pi int, s snapshot) error
	}

	repository struct {
		db     db.IDatabase
		logger slog.INamedLogger
	}
)

func newRepository(logger slog.ILogger, db db.IDatabase) IRepository {
	return &repository{
		db:     db,
		logger: slog.NewNamed(logger, "diagnostics-repository", slog.None),
	}
}

func (r *repository) FindOne(ctx context.Context, pkg string, pi int) (*snapshot, error) {
	const q = `
SELECT last_modified, 
       hash 
FROM snapshots 
WHERE (package, property_index) = ($1, $2)
ORDER BY created_at
LIMIT 1;
`
	rows, err := r.db.QueryContext(ctx, q, pkg, pi)
	if err != nil {
		return nil, errors.Wrapf(err, "diagnostics: failed to query snapshot for pkg=%s, i=%d", pkg, pi)
	}
	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)

	var s snapshot
	for rows.Next() {
		if err = rows.Scan(&s.LastModified, &s.Hash); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func (r *repository) InsertOne(ctx context.Context, pkg string, pi int, s snapshot) error {
	const q = `
INSERT into snapshots (package, 
					   property_index, 
					   last_modified, 
					   hash) 
VALUES ($1, $2, $3, $4);
`
	_, err := r.db.ExecContext(ctx, q, pkg, pi, s.LastModified, s.Hash)
	if err != nil {
		return errors.Wrapf(err, "diagnostics: failed to insert snapshot for pkg=%s, i=%d", pkg, pi)
	}
	return nil
}

func (r *repository) SeedDB() error {
	log := func(msg string) {
		r.logger.Log("seed", "msg", msg)
	}

	if !r.db.RequiresSeed() {
		log("seeding not required")
		return nil
	}

	log("seeding database")

	qs := []string{
		`
		CREATE TABLE IF NOT EXISTS runs (
		   id INTEGER PRIMARY KEY AUTOINCREMENT,
           arguments JSON NOT NULL,
		   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`
		CREATE TABLE IF NOT EXISTS snapshots (
		   id INTEGER PRIMARY KEY AUTOINCREMENT,
		   run_id INTEGER NOT NULL,
		   package TEXT NOT NULL,
		   property_index INTEGER NOT NULL,
		   last_modified INTEGER NOT NULL,
		   hash INTEGER NOT NULL,
           created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_package_property_index ON snapshots (package, property_index);`,
	}

	tx, err := r.db.Conn().Begin()
	if err != nil {
		return errors.Wrap(err, "diagnostics: failed to begin transaction")
	}
	for i, q := range qs {
		if _, err = tx.ExecContext(context.Background(), q); err != nil {
			if err = tx.Rollback(); err != nil {
				return errors.Wrap(err, "diagnostics: failed to rollback transaction")
			}
			return errors.Wrapf(err, "diagnostics: failed to seed database: stmt[%d]", i)
		}
	}
	if err = tx.Commit(); err != nil {
		err = errors.Wrap(err, "diagnostics: failed to commit transaction")
	}
	return err
}
