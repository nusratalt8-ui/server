package dbutil

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
	cfg Config
}

func New(cfg Config) (*DB, error) {
	if dir := filepath.Dir(cfg.Path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, errors.Join(ErrOpenFailed, err)
		}
	}
	dsn := cfg.Path +
		"?_pragma=busy_timeout(" + strconv.Itoa(cfg.BusyTimeout) + ")" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=foreign_keys(ON)"
	handle, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.Join(ErrOpenFailed, err)
	}
	if cfg.MaxOpenConns > 0 {
		handle.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if err := handle.Ping(); err != nil {
		return nil, errors.Join(ErrPingFailed, err)
	}
	db := &DB{sql: handle, cfg: cfg}
	if err := db.applySchema(); err != nil {
		handle.Close()
		return nil, err
	}
	return db, nil
}

func (d *DB) applySchema() error {
	if d.cfg.SchemaPath == "" {
		return nil
	}
	data, err := os.ReadFile(d.cfg.SchemaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.Join(ErrSchemaFailed, err)
	}
	if _, err := d.sql.Exec(string(data)); err != nil {
		return errors.Join(ErrSchemaFailed, err)
	}
	// additive migrations — safe to run on every boot
	migrations := []string{
		`ALTER TABLE agent_key ADD COLUMN key_raw TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN theme_id TEXT NOT NULL DEFAULT 'dark'`,
		`ALTER TABLE users ADD COLUMN ui_prefs TEXT NOT NULL DEFAULT '{}'`,
		`ALTER TABLE users ADD COLUMN plan INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		d.sql.Exec(m) // ignore error — column already exists on fresh DBs
	}
	return nil
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.sql.Exec(query, args...)
}

func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.sql.Query(query, args...)
}

func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.sql.QueryRow(query, args...)
}

func (d *DB) Begin() (*sql.Tx, error) {
	return d.sql.Begin()
}

func (d *DB) Close() error {
	return d.sql.Close()
}
