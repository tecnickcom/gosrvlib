/*
Package sqlconn provides a simple way to connect to a SQL database and manage
the connection. It is based on the database/sql package and provides a way to
perform a health check on the connection. This packages also provide a way to
gracefully shutdown the connection when the application is shutting down.
*/
package sqlconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Vonage/gosrvlib/pkg/logging"
	"go.uber.org/zap"
)

// ConnectFunc is the type of function called to perform the actual DB connection.
type ConnectFunc func(ctx context.Context, cfg *config) (*sql.DB, error)

// CheckConnectionFunc is the type of function called to perform a DB connection check.
type CheckConnectionFunc func(ctx context.Context, db *sql.DB) error

// SQLOpenFunc is the type of function called to open the DB. (Only for monkey patch testing).
type SQLOpenFunc func(driverName, dataSourceName string) (*sql.DB, error)

// SQLConn is the structure that helps to manage a SQL DB connection.
type SQLConn struct {
	cfg    *config
	db     *sql.DB
	dbLock sync.RWMutex
	logger *zap.Logger
}

// Connect attempts to connect to a SQL database.
func Connect(ctx context.Context, url string, opts ...Option) (*SQLConn, error) {
	driver, dsn, err := parseConnectionURL(url)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig(driver, dsn)

	for _, applyOpt := range opts {
		applyOpt(cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	db, err := cfg.connectFunc(ctx, cfg)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	db.SetConnMaxLifetime(cfg.connMaxLifetime)
	db.SetMaxIdleConns(cfg.connMaxIdleCount)
	db.SetMaxOpenConns(cfg.connMaxOpenCount)

	l := logging.WithComponent(ctx, "sqlconn")

	c := SQLConn{
		cfg:    cfg,
		db:     db,
		logger: l,
	}

	// wait for shutdown signal or context cancelation
	go func() {
		select {
		case <-cfg.shutdownSignalChan:
			l.Debug("sqlconn shutdown signal received")
		case <-ctx.Done():
			l.Warn("sqlconn context canceled")
		}

		_ = c.Shutdown(ctx)
	}()

	cfg.shutdownWaitGroup.Add(1)

	return &c, nil
}

// DB returns a database connection from the pool.
func (c *SQLConn) DB() *sql.DB {
	c.dbLock.RLock()
	defer c.dbLock.RUnlock()

	return c.db
}

// HealthCheck performs a health check of the database connection.
func (c *SQLConn) HealthCheck(ctx context.Context) error {
	c.dbLock.RLock()
	defer c.dbLock.RUnlock()

	if c.db == nil {
		return errors.New("database not unavailable")
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.pingTimeout)
	defer cancel()

	return c.cfg.checkConnectionFunc(ctx, c.db)
}

// Shutdown closes the database and prevents new queries from starting.
// It waits for all queries that have started processing on the server to finish.
func (c *SQLConn) Shutdown(_ context.Context) error {
	c.logger.Debug("shutting down sql connection")

	c.dbLock.Lock()
	defer c.dbLock.Unlock()

	err := c.db.Close()

	c.db = nil
	c.cfg.shutdownWaitGroup.Add(-1)

	c.logger.Debug("sql connection shutdown complete", zap.Error(err))

	return err //nolint:wrapcheck
}

func checkConnection(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed ping on database: %w", err)
	}

	//nolint:rowserrcheck
	rows, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		return fmt.Errorf("failed running check query on database: %w", err)
	}

	defer logging.Close(ctx, rows, "failed closing SQL rows")

	return nil
}

func connectWithBackoff(ctx context.Context, cfg *config) (*sql.DB, error) {
	db, err := cfg.sqlOpenFunc(cfg.driver, cfg.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.pingTimeout)
	defer cancel()

	if err = cfg.checkConnectionFunc(ctx, db); err != nil {
		return nil, fmt.Errorf("failed checking database connection: %w", err)
	}

	return db, nil
}

// parseConnectionURL attempts to extract the driver/dsn pair from a string in the format <DRIVER>://<DSN>
// if only the DSN part is set, the driver will need to be specified via a configuration option.
func parseConnectionURL(url string) (string, string, error) {
	if url == "" {
		return "", "", nil
	}

	parts := strings.Split(url, "://")

	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid connection string: %q", url)
}
