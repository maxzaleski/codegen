package diagnostics

import (
	"context"
	"database/sql"
	"github.com/maxzaleski/codegen/internal/core/slog"
)

func Run(ctx context.Context, l slog.ILogger, codegenLoc string) error {
	// 1. Open database connection.
	db, err := newDB(l, codegenLoc)
	if err != nil {
		return err
	}
	defer func(conn *sql.DB) { _ = conn.Close() }(db.Conn())

	// 2.
	return nil
}
