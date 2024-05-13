package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Statistic struct {
	Domain               string              `json:"domain"`
	StartTime            *time.Time          `json:"start_time"`
	EndTime              *time.Time          `json:"end_time"`
	CurrentVisitors      int                 `json:"current_visitors"`
	TotalPageViews       int                 `json:"total_page_views"`
	TotalVisitors        int                 `json:"total_visitors"`
	PageViewsPerHour     *map[string]int32   `json:"page_views_per_hour"`
	QuickSyncsPerHour    *map[string]int32   `json:"quick_syncs_per_hour"`
	VisitorsPerCountry   *map[string]int32   `json:"visitors_per_country"`
	RequestsPerIp        *[]requestsPerIp    `json:"requests_per_ip"`
	Referrers            *map[string]int32   `json:"referrers"`
	VisitorsPerUtmSource *map[string]int32   `json:"visitors_per_utm_source"`
	RevenuePerUtmSource  *map[string]float32 `json:"revenue_per_utm_source"`
	RevenuePerReferrer   *map[string]float32 `json:"revenue_per_referrer"`
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
}

type request struct {
	Domain      string
	Duration    int64
	Timestamp   time.Time
	UserAgent   string
	Referrer    *string
	Path        string
	QueryParams *map[string]interface{}
	Country     string
	StatusCode  int16
	Ip          net.IP
	Ips         *[]net.IP
}

type requestsPerIp struct {
	Ip      net.IP    `json:"ip"`
	Ips     *[]net.IP `json:"ips"`
	Country string    `json:"country"`
	Count   int32     `json:"count"`
}

func getTotalPageViews(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(*) AS c FROM public.events WHERE event_name = 'pageview' AND domain = $1"

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
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for total page views")
		return 0, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan page view rows")
		return 0, err
	}

	return result, nil
}

func getTotalVisitors(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(DISTINCT visitor_id) AS c FROM public.events WHERE event_name = 'pageview' AND domain = $1"

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
	var query = "SELECT COUNT(DISTINCT visitor_id) AS c FROM public.events WHERE event_name = 'pageview' AND domain = $1 AND timestamp > $2"

	var start = time.Now().UTC().Add(time.Duration(-5) * time.Minute)
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
	var query = "SELECT domain, event_name, duration, timestamp, user_agent, referrer, path, visitor_id, query_params, country, event_data, status_code FROM public.events WHERE domain = $1"

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
		var duration sql.NullInt64

		err := rows.Scan(&e.Domain, &e.EventName, &duration, &e.Timestamp, &e.UserAgent, &e.Referrer, &e.Path, &e.VisitorId, &queryJson, &e.Country, &eventJson, &e.StatusCode)

		if err != nil {
			return nil, err
		}

		if duration.Valid {
			e.Duration = duration.Int64
		} else {
			e.Duration = 0
		}

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

func getRequests(db *sql.DB, domain string, start *time.Time, end *time.Time) (*[]request, error) {
	var query = "SELECT domain, duration, timestamp, user_agent, referrer, path, query_params, country, status_code, ip, ips FROM public.monthly_traffic WHERE domain = $1"

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
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for requests")
		return nil, err
	}

	var result = make([]request, 0)

	for rows.Next() {
		var e request
		var queryJson sql.NullString
		var duration sql.NullInt64
		var ip string
		ips := make([]string, 10)

		err := rows.Scan(&e.Domain, &duration, &e.Timestamp, &e.UserAgent, &e.Referrer, &e.Path, &queryJson, &e.Country, &e.StatusCode, &ip, pq.Array(&ips))

		if err != nil {
			return nil, err
		}

		e.Ip = net.ParseIP(ip)

		if duration.Valid {
			e.Duration = duration.Int64
		} else {
			e.Duration = 0
		}

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

		result = append(result, e)
	}

	return &result, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func groupRequestsPerIp(requests *[]request) (*[]requestsPerIp, error) {
	eventsPerPath := make(map[string]*requestsPerIp)

	for _, e := range *requests {
		key := e.Ip.String()

		_, ok := eventsPerPath[key]

		if !ok {
			x := requestsPerIp{
				Count:   1,
				Ip:      e.Ip,
				Ips:     e.Ips,
				Country: e.Country,
			}
			eventsPerPath[key] = &x
		} else {
			eventsPerPath[key].Count += 1
		}
	}

	result := make([]requestsPerIp, 0)

	for _, v := range eventsPerPath {
		result = append(result, *v)
	}

	return &result, nil
}

func getOriginalReferringDomain(db *sql.DB, visitorId string, domain string) (string, error) {
	var query = "SELECT referrer FROM public.events WHERE visitor_id = $1 ORDER BY timestamp ASC LIMIT 1"
	rows, err := db.Query(query, visitorId)
	var result = ""

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for original referrer")
		return result, err
	}

	var referrer sql.NullString
	rows.Next()
	err = rows.Scan(&referrer)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan original referrer")
		return result, err
	}

	if referrer.Valid {

		u, err := url.Parse(referrer.String)
		if err == nil {
			if u.Host != domain && u.Host != "www."+domain {
				result = u.Host
			}
		}
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
	stats.StartTime = start
	stats.EndTime = end

	visits, err := getTotalPageViews(db, domain, start, end)
	if err != nil {
		return nil, err
	}
	stats.TotalPageViews = visits

	visitors, err := getTotalVisitors(db, domain, start, end)
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

	pageViewsPerHour := make(map[string]int32)
	quickSyncsPerHour := make(map[string]int32)
	visitorsPerCountry := make(map[string]int32)
	pageViewsPerReferrer := make(map[string]int32)
	visitorsPerUtmSource := make(map[string]int32)
	revenuePerUtmSource := make(map[string]float32)
	revenuePerReferrer := make(map[string]float32)
	visitorIds := make([]string, 0)
	utmSourceVisitors := make([]string, 0)

	for _, e := range *events {
		if e.EventName == "pageview" {

			// group pageviews
			key := e.Timestamp.Format("2006-01-02 15")
			_, ok := pageViewsPerHour[key]

			if !ok {
				pageViewsPerHour[key] = 1
			} else {
				pageViewsPerHour[key] += 1
			}

			// group visitors per country
			if stringInSlice(e.VisitorId, visitorIds) {
				continue
			}

			_, ok = visitorsPerCountry[e.Country]

			if !ok {
				visitorsPerCountry[e.Country] = 1
			} else {
				visitorsPerCountry[e.Country] += 1
			}

			visitorIds = append(visitorIds, e.VisitorId)

			// group page views per referrer
			var referrer = ""

			if e.Referrer != nil && len(*e.Referrer) > 0 {
				u, err := url.Parse(*e.Referrer)
				if err != nil {
					continue
				}

				if u.Host != domain {
					referrer = u.Host
				}
			}

			if len(referrer) > 0 {
				_, ok = pageViewsPerReferrer[referrer]

				if !ok {
					pageViewsPerReferrer[referrer] = 1
				} else {
					pageViewsPerReferrer[referrer] += 1
				}
			}

			// group page views per utm source
			if !stringInSlice(e.VisitorId, utmSourceVisitors) && e.QueryParams != nil {
				key, ok := (*e.QueryParams)["utm_source"].(string)

				if ok {
					_, ok = visitorsPerUtmSource[key]

					if !ok {
						visitorsPerUtmSource[key] = 1
					} else {
						visitorsPerUtmSource[key] += 1
					}

					utmSourceVisitors = append(visitorIds, e.VisitorId)
				}
			}

			// revenue per utm source and referrer
			if e.QueryParams != nil {
				sale, ok := (*e.QueryParams)["sale_total"].(string)

				if ok {
					f, err := strconv.ParseFloat(sale, 32)

					if err == nil {

						source, ok := (*e.QueryParams)["utm_source"].(string)

						if ok {

							_, sok := revenuePerUtmSource[source]

							if !sok {
								revenuePerUtmSource[source] = 0
							}

							revenuePerUtmSource[source] += float32(f)
						}

						referrer, err := getOriginalReferringDomain(db, e.VisitorId, domain)

						if err == nil {
							if len(referrer) > 0 {
								_, rok := revenuePerReferrer[referrer]

								if !rok {
									revenuePerReferrer[referrer] = 0
								}

								revenuePerReferrer[referrer] += float32(f)
							}
						}
					}
				}
			}
		} else if e.EventName == "quicksync" {
			key := e.Timestamp.Format("2006-01-02 15")

			val, ok := quickSyncsPerHour[key]

			if !ok {
				val = 1
				quickSyncsPerHour[key] = val
			} else {
				quickSyncsPerHour[key] = val + 1
			}
		}
	}

	stats.PageViewsPerHour = &pageViewsPerHour
	stats.QuickSyncsPerHour = &quickSyncsPerHour
	stats.VisitorsPerCountry = &visitorsPerCountry
	stats.Referrers = &pageViewsPerReferrer
	stats.VisitorsPerUtmSource = &visitorsPerUtmSource
	stats.RevenuePerUtmSource = &revenuePerUtmSource
	stats.RevenuePerReferrer = &revenuePerReferrer

	// requests
	req, err := getRequests(db, domain, start, end)

	if err != nil {
		return nil, err
	}

	rpi, err := groupRequestsPerIp(req)

	if err != nil {
		return nil, err
	}

	stats.RequestsPerIp = rpi

	return &stats, nil
}
