package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Statistic struct {
	Domain          string            `json:"domain"`
	StartTime       *time.Time        `json:"start_time"`
	EndTime         *time.Time        `json:"end_time"`
	CurrentVisitors int               `json:"current_visitors"`
	TotalPageViews  int               `json:"total_page_views"`
	TotalVisitors   int               `json:"total_visitors"`
	RequestsPerHour *map[string]int32 `json:"requests_per_hour"`
	RequestsPerIp   *map[string]int32 `json:"requests_per_ip"`
}

type event struct {
	Domain      string
	EventName   string
	Duration    int64
	Timestamp   time.Time
	UserAgent   string
	Referrer    *string
	VisitorId   string
	Path        string
	QueryParams *map[string]interface{}
	Country     string
	EventData   *map[string]interface{}
	StatusCode  int16
	Ip          net.IP
}

func getTotalVisits(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(*) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1"

	if start != nil {
		query = query + " AND timestamp >= $2"
	}

	if end != nil {
		if start != nil {
			query = query + " AND timestamp::date <= $3"
		} else {
			query = query + " AND timestamp::date <= $2"
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
	var query = "SELECT COUNT(DISTINCT ip) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1"

	if start != nil {
		query = query + " AND timestamp::date >= $2"
	}

	if end != nil {
		if start != nil {
			query = query + " AND timestamp::date <= $3"
		} else {
			query = query + " AND timestamp::date <= $2"
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
	var query = "SELECT COUNT(DISTINCT ip) AS c FROM public.traffic WHERE event_name = 'pageview' AND domain = $1 AND timestamp > $2"

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

func getAllEvents(db *sql.DB, domain string, start *time.Time, end *time.Time) (*[]event, error) {
	var query = "SELECT domain, event_name, duration, timestamp, user_agent, referrer, path, visitor_id, query_params, country, event_data, status_code, ip FROM public.traffic WHERE domain = $1"

	if start != nil {
		query = query + " AND timestamp::date >= $2"
	}

	if end != nil {
		if start != nil {
			query = query + " AND timestamp::date <= $3"
		} else {
			query = query + " AND timestamp::date <= $2"
		}
	}

	query += " ORDER BY timestamp"

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
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for all events")
		return nil, err
	}

	var result = make([]event, 0)

	for rows.Next() {
		var e event
		var queryJson sql.NullString
		var eventJson sql.NullString
		var ip string

		err := rows.Scan(&e.Domain, &e.EventName, &e.Duration, &e.Timestamp, &e.UserAgent, &e.Referrer, &e.Path, &e.VisitorId, &queryJson, &e.Country, &eventJson, &e.StatusCode, &ip)

		if err != nil {
			return nil, err
		}

		e.Ip = net.ParseIP(ip)

		if //goland:noinspection GoDfaConstantCondition
		queryJson.Valid {
			q := make(map[string]interface{})

			err := json.Unmarshal([]byte(queryJson.String), &q)

			if err != nil {
				log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to unmarshal query json")
				return nil, err
			}

			e.QueryParams = &q
		}

		if //goland:noinspection GoDfaConstantCondition
		eventJson.Valid {
			q := make(map[string]interface{})

			err := json.Unmarshal([]byte(eventJson.String), &q)

			if err != nil {
				log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to unmarshal event json")
				return nil, err
			}

			e.EventData = &q
		}

		result = append(result, e)
	}

	return &result, nil
}

func groupEventsPerHour(events *[]event) (*map[string]int32, error) {
	eventsPerHour := make(map[string]int32)

	for _, e := range *events {
		key := e.Timestamp.Format("2006-01-02 15")

		val, ok := eventsPerHour[key]

		if !ok {
			val = 1
			eventsPerHour[key] = val
		} else {
			eventsPerHour[key] = val + 1
		}
	}

	return &eventsPerHour, nil
}

func groupEventsPerIp(events *[]event) (*map[string]int32, error) {
	eventsPerIp := make(map[string]int32)

	for _, e := range *events {
		key := e.Ip.String()

		val, ok := eventsPerIp[key]

		if !ok {
			val = 1
			eventsPerIp[key] = val
		} else {
			eventsPerIp[key] = val + 1
		}
	}

	return &eventsPerIp, nil
}

func GetStats(domain string, start *time.Time, end *time.Time) (*Statistic, error) {
	db, err := sql.Open("postgres", ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var stats Statistic
	stats.Domain = domain
	stats.StartTime = start
	stats.EndTime = end

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

	events, err := getAllEvents(db, domain, start, end)

	if err != nil {
		return nil, err
	}

	eph, err := groupEventsPerHour(events)

	if err != nil {
		return nil, err
	}

	stats.RequestsPerHour = eph

	epi, err := groupEventsPerIp(events)

	if err != nil {
		return nil, err
	}

	stats.RequestsPerIp = epi

	return &stats, nil
}
