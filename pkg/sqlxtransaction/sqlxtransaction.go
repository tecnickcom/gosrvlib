/*
Package sqlxtransaction provides a simple way to execute a function inside an
SQLX transaction. The function to be executed is passed as an argument to the
Exec function.

For a similar functionality using the standard database/sql package instead of
the github.com/jmoiron/sqlx one, see the
github.com/Vonage/gosrvlib/pkg/sqltransaction package.
*/
package sqlxtransaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Vonage/gosrvlib/pkg/logging"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ExecFunc is the type of the function to be executed inside a SQL Transaction.
type ExecFunc func(ctx context.Context, tx *sqlx.Tx) error

// DB is the interface which represents the database driver.
type DB interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Exec executes the specified function inside a SQL transaction.
func Exec(ctx context.Context, db DB, run ExecFunc) error {
	return ExecWithOptions(ctx, db, run, nil)
}

// ExecWithOptions executes the specified function inside a SQL transaction.
func ExecWithOptions(ctx context.Context, db DB, run ExecFunc, opts *sql.TxOptions) error {
	var committed bool

	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("unable to start SQLX transaction: %w", err)
	}

	defer func() {
		if committed {
			return
		}

		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			logging.FromContext(ctx).Error("failed rolling back SQLX transaction", zap.Error(err))
		}
	}()

	if err = run(ctx, tx); err != nil {
		return fmt.Errorf("failed executing a function inside SQLX transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit SQL transaction: %w", err)
	}

	committed = true

	return nil
}
