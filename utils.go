package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "..")
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

func IpBetween(from net.IP, to net.IP, test net.IP) (bool, error) {
	if from == nil || to == nil || test == nil {
		return false, errors.New("an input is nil")
	}

	from16 := from.To16()
	to16 := to.To16()
	test16 := test.To16()
	if from16 == nil || to16 == nil || test16 == nil {
		return false, errors.New("an ip did not convert to a 16 byte")
	}

	if bytes.Compare(test16, from16) >= 0 && bytes.Compare(test16, to16) <= 0 {
		return true, nil
	}

	return false, nil
}

type IpRangeCountry struct {
	from    net.IP
	to      net.IP
	country string
}

func ReadDbIpCsv() (*[]IpRangeCountry, error) {
	file, err := os.Open("dbip-country-lite.csv")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	result := make([]IpRangeCountry, 0)

	for _, record := range records {
		from := net.ParseIP(record[0])
		to := net.ParseIP(record[1])
		country := record[2]

		result = append(result, IpRangeCountry{from: from, to: to, country: country})
	}

	return &result, nil
}
