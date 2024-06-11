package main

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b))
)

func MigrateDb(db *sql.DB) {
	createMigrationsTable(db)
	currentVersion := getCurrentMigratedVersion(db)
	availableVersions := getAvailableMigrations(currentVersion)

	var appliedVersion = currentVersion

	for _, filePath := range *availableVersions {
		c, err := os.ReadFile(filePath)
		if err != nil {
			log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatalf("Error reading migration file %s", filePath)
		}

		script := string(c)
		_, err = db.Exec(script)
		if err != nil {
			log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatalf("Failed to apply migration file %s", filePath)
		}

		y := strings.Split(filePath, "/")
		x, err := getVersionFromFileName(y[len(y)-1])

		if err != nil {
			log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatalf("Failed to get version from filename %s", filePath)
		}

		appliedVersion = x
	}

	if appliedVersion != currentVersion {
		writeCurrentVersion(db, appliedVersion)
	}
}

func getVersionFromFileName(fileName string) (int16, error) {
	prefix := strings.Split(fileName, "_")[0]

	nr, err := strconv.Atoi(prefix)

	if err == nil {
		return int16(nr), nil
	} else {
		return -1, err
	}
}

func createMigrationsTable(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS migration_history (version SMALLINT PRIMARY KEY, migrated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)")

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatal("Failed to create migrations history table")
	}
}

func getCurrentMigratedVersion(db *sql.DB) int16 {

	var version int16
	err := db.QueryRow("SELECT version FROM migration_history ORDER BY migrated DESC LIMIT 1").Scan(&version)

	if errors.Is(err, sql.ErrNoRows) {
		return 0
	} else if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatal("Failed to get latest migration version")
	}

	return version
}

func getAvailableMigrations(currentVersion int16) *[]string {
	path := filepath.Join(Root, "migrations")
	files, err := os.ReadDir(path)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatal("Failed to get available migrations")
	}

	var result = make([]string, 0)

	for _, file := range files {
		nr, atoiErr := getVersionFromFileName(file.Name())

		if atoiErr == nil {
			if nr > currentVersion {
				result = append(result, filepath.Join(path, file.Name()))
			} else {
				log.WithFields(log.Fields{"currentVersion": currentVersion}).Debugf("Skipping migration %s", file.Name())
			}
		} else {
			log.WithFields(log.Fields{"file": file.Name()}).Fatalf("Failed to get version from file %s", file.Name())
		}
	}

	sort.Strings(result)

	return &result
}

func writeCurrentVersion(db *sql.DB, currentVersion int16) {
	_, err := db.Exec("INSERT INTO migration_history (version) VALUES ($1)", currentVersion)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Fatalf("Failed to write migration version")
	}
}
