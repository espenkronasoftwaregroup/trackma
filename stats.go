package main

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type Statistic struct {
	Domain          string     `json:"domain"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	CurrentVisitors int        `json:"current_visitors"`
	TotalPageViews  int        `json:"total_page_views"`
	TotalVisitors   int        `json:"total_visitors"`
}

func getTotalVisits(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(*) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1"

	if start != nil {
		query = query + " AND timestamp > $2"
	}

	if end != nil {
		if start != nil {
			query = query + " AND timestamp < $3"
		} else {
			query = query + " AND timestamp < $2"
		}
	}

	var rows *sql.Rows
	var err error

	if start != nil && end != nil {
		rows, err = db.Query(query, domain, start, end)
	} else if start != nil {
		rows, err = db.Query(query, domain, start)
	} else if end != nil {
		rows, err = db.Query(query, domain, end)
	} else {
		rows, err = db.Query(query, domain)
	}

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for total visits")
		return 0, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan rows")
		return 0, err
	}

	return result, nil
}

func getVisitors(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(DISTINCT visitor_id) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1"

	if start != nil {
		query = query + " AND timestamp > $2"
	}

	if end != nil {
		if start != nil {
			query = query + " AND timestamp < $3"
		} else {
			query = query + " AND timestamp < $2"
		}
	}

	var rows *sql.Rows
	var err error

	if start != nil && end != nil {
		rows, err = db.Query(query, domain, start, end)
	} else if start != nil {
		rows, err = db.Query(query, domain, start)
	} else if end != nil {
		rows, err = db.Query(query, domain, end)
	} else {
		rows, err = db.Query(query, domain)
	}

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for visitors")
		return 0, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan rows")
		return 0, err
	}

	return result, nil
}

func getCurrentVisitors(db *sql.DB, domain string) (int, error) {
	var query = "SELECT COUNT(DISTINCT visitor_id) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1 AND timestamp > $2"

	var start = time.Now().Add(time.Duration(-5) * time.Minute)
	rows, err := db.Query(query, domain, start)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for current visitors")
		return 0, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan rows")
		return 0, err
	}

	return result, nil
}

func GetStats(domain string, start *time.Time, end *time.Time) (*Statistic, error) {
	db, err := sql.Open("postgres", ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var stats Statistic
	stats.Domain = domain

	visits, err := getTotalVisits(db, domain, start, end)
	if err != nil {
		return nil, err
	}
	stats.TotalPageViews = visits

	visitors, err := getVisitors(db, domain, start, end)
	if err != nil {
		return nil, err
	}
	stats.TotalVisitors = visitors

	currentVisitors, err := getCurrentVisitors(db, domain)
	if err != nil {
		return nil, err
	}
	stats.CurrentVisitors = currentVisitors

	return &stats, nil
}
