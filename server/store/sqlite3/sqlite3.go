package sqlite3

import (
	"fmt"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
)

type SQLite3 struct {
	*sqlx.DB
}

func New(path string) (*SQLite3, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	sdb := &SQLite3{DB: db}
	return sdb, sdb.init()
}

func (sdb SQLite3) init() error {
	var version int
	if err := sdb.Get(&version, `PRAGMA user_version`); err != nil {
		return fmt.Errorf("getting database version: %w", err)
	}
	if version != 1 {
		if _, err := sdb.Exec(schema1); err != nil {
			return fmt.Errorf("initializing database: %w", err)
		}
	}
	return nil
}
