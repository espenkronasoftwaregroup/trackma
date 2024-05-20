package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
	"net/url"
	"sort"
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
	PageViewsPerHour     *map[string]*int32  `json:"page_views_per_hour"`
	QuickSyncsPerHour    *map[string]*int32  `json:"quick_syncs_per_hour"`
	VisitorsPerCountry   *map[string]*int32  `json:"visitors_per_country"`
	RequestsPerIp        *[]requestsPerIp    `json:"requests_per_ip"`
	Referrers            *map[string]*int32  `json:"referrers"`
	VisitorsPerUtmSource *map[string]*int32  `json:"visitors_per_utm_source"`
	RevenuePerUtmSource  *map[string]float32 `json:"revenue_per_utm_source"`
	RevenuePerReferrer   *map[string]float32 `json:"revenue_per_referrer"`
}

type event struct {
	//Id          int64
	Domain      string
	EventName   string
	Duration    int64
	Timestamp   time.Time
	UserAgent   string
	Referrer    *string
	VisitorId   string
	SessionId   string
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
		return -1, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)
	rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan page view rows")
		return -1, err
	}

	return result, nil
}

func getTotalVisitors(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(DISTINCT session_id) AS c FROM public.events WHERE event_name = 'pageview' AND session_id IS NOT NULL AND domain = $1"

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
	rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan rows")
		return 0, err
	}

	return result, nil
}

func getCurrentVisitors(db *sql.DB, domain string) (int, error) {
	var query = "SELECT COUNT(DISTINCT session_id) AS c FROM public.events WHERE event_name = 'pageview' AND session_id IS NOT NULL AND domain = $1 AND timestamp > $2"

	var start = time.Now().UTC().Add(time.Duration(-5) * time.Minute)
	rows, err := db.Query(query, domain, start)

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for current visitors")
		return 0, err
	}

	var result int = 0

	rows.Next()
	err = rows.Scan(&result)
	rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan rows")
		return 0, err
	}

	return result, nil
}

func countEvents(db *sql.DB, domain string, start *time.Time, end *time.Time) (int, error) {
	var query = "SELECT COUNT(*) FROM public.events WHERE domain = $1"

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
		return 0, err
	}

	var result = 0
	rows.Next()
	err = rows.Scan(&result)
	rows.Close()

	if err != nil {
		return 0, err
	}

	return result, nil
}

func getAllEvents(db *sql.DB, c chan *event, domain string, start *time.Time, end *time.Time) {

	var pageSize = 10000
	eventCount, err := countEvents(db, domain, start, end)
	var pageCount = math.Ceil(float64(eventCount) / float64(pageSize))
	var lastStart time.Time

	if err != nil {
		log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to count events")
		close(c)
	}

	for i := 0; i <= int(pageCount); i++ {
		var query = "SELECT domain, event_name, duration, timestamp, user_agent, referrer, path, session_id, visitor_id, query_params, country, event_data, status_code FROM public.events WHERE domain = $1"

		if !lastStart.IsZero() {
			query = query + " AND timestamp >= $2"
		} else if start != nil {
			query = query + " AND timestamp::date >= $2"
		}

		if end != nil {
			if !lastStart.IsZero() || start != nil {
				query = query + " AND timestamp::date <= $3"
			} else {
				query = query + " AND timestamp::date <= $2"
			}
		}

		query += " ORDER BY timestamp ASC LIMIT 10000" // it would be correct to use pageSize here but im lazy

		var rows *sql.Rows

		if !lastStart.IsZero() && end != nil {
			rows, err = db.Query(query, domain, lastStart, end)
		} else if start != nil && end != nil {
			rows, err = db.Query(query, domain, start, end)
		} else if !lastStart.IsZero() {
			rows, err = db.Query(query, domain, lastStart)
		} else if end != nil {
			rows, err = db.Query(query, domain, end)
		} else {
			rows, err = db.Query(query, domain)
		}

		if err != nil {
			log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to query for all events")
			close(c)
		}

		for rows.Next() {
			var e event
			var queryJson sql.NullString
			var eventJson sql.NullString
			var sessionId sql.NullString
			var duration sql.NullInt64

			err := rows.Scan(&e.Domain, &e.EventName, &duration, &e.Timestamp, &e.UserAgent, &e.Referrer, &e.Path, &sessionId, &e.VisitorId, &queryJson, &e.Country, &eventJson, &e.StatusCode)

			if err != nil {
				log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to scan events")
				close(c)
			}

			if duration.Valid {
				e.Duration = duration.Int64
			} else {
				e.Duration = 0
			}

			if sessionId.Valid {
				e.SessionId = sessionId.String
			}

			if //goland:noinspection GoDfaConstantCondition
			queryJson.Valid {
				q := make(map[string]interface{})

				err := json.Unmarshal([]byte(queryJson.String), &q)

				if err != nil {
					log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to unmarshal query json")
					close(c)
				}

				e.QueryParams = &q
			}

			if //goland:noinspection GoDfaConstantCondition
			eventJson.Valid {
				q := make(map[string]interface{})

				err := json.Unmarshal([]byte(eventJson.String), &q)

				if err != nil {
					log.WithFields(log.Fields{"error": fmt.Errorf("%w", err)}).Error("Failed to unmarshal event json")
					close(c)
				}

				e.EventData = &q
			}

			c <- &e

			lastStart = e.Timestamp
		}

		rows.Close()
	}

	close(c)
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

	rows.Close()

	return &result, nil
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

	type kv struct {
		Key   string
		Value requestsPerIp
	}

	values := make([]kv, len(eventsPerPath))

	count := 0
	for k, v := range eventsPerPath {
		values[count] = kv{
			k, *v,
		}
		count++
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Value.Count > values[j].Value.Count
	})

	requestCount := 10

	if len(values) < requestCount {
		requestCount = len(values)
	}

	result := make([]requestsPerIp, requestCount)

	for i, v := range values[:requestCount] {
		result[i] = v.Value
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
	rows.Close()

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

func increment(counts map[string]*int32, key string) {
	if p, ok := counts[key]; ok {
		*p++
		return
	}

	var n int32 = 1
	counts[key] = &n
}

func GetStats(db *sql.DB, domain string, start *time.Time, end *time.Time) (*Statistic, error) {
	var readChannel = make(chan *event, 100000)

	var stats Statistic
	stats.Domain = domain
	stats.StartTime = start
	stats.EndTime = end

	tpv, err := getTotalPageViews(db, domain, start, end)
	if err != nil {
		return nil, err
	}
	stats.TotalPageViews = tpv

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

	go getAllEvents(db, readChannel, domain, start, end)

	pageViewsPerHour := make(map[string]*int32)
	quickSyncsPerHour := make(map[string]*int32)
	visitorsPerCountry := make(map[string]*int32)
	pageViewsPerReferrer := make(map[string]*int32)
	visitorsPerUtmSource := make(map[string]*int32)
	revenuePerUtmSource := make(map[string]float32)
	revenuePerReferrer := make(map[string]float32)
	visitorIds := make(map[string]bool)
	utmSourceVisitors := make(map[string]bool)

	for e := range readChannel {
		if e.EventName == "pageview" {

			// group pageviews
			key := e.Timestamp.String()[:13]
			increment(pageViewsPerHour, key)

			// group visitors per country
			_, exists := visitorIds[e.VisitorId]
			if !exists {
				increment(visitorsPerCountry, e.Country)
				visitorIds[e.VisitorId] = true
			}

			// group page views per referrer
			referrer := ""

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
				increment(pageViewsPerReferrer, referrer)
			}

			// group page views per utm source
			_, ok := utmSourceVisitors[e.VisitorId]
			if !ok && e.QueryParams != nil {
				key, ok := (*e.QueryParams)["utm_source"].(string)

				if ok {
					increment(visitorsPerUtmSource, key)
					utmSourceVisitors[e.VisitorId] = true
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
			key := e.Timestamp.String()[:13]
			increment(quickSyncsPerHour, key)
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
