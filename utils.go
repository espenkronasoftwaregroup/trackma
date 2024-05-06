package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b))
)

func MigrateDb(db *sql.DB) error {
	path := filepath.Join(Root, "migrations", "001_init.sql")

	c, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	script := string(c)
	_, err = db.Exec(script)
	if err != nil {
		return err
	}

	return nil
}
