package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IngestRequest struct {
	Domain          string   `json:"domain"`
	Path            string   `json:"path"`
	Query           string   `json:"query"`
	EventName       string   `json:"eventName"`
	VisitorId       string   `json:"visitorId"`
	Referrer        string   `json:"referrer"`
	ClientIp        []string `json:"clientIp"`
	ClientUserAgent string   `json:"clientUserAgent"`
	Duration        int64    `json:"duration"`
	StatusCode      int16    `json:"statusCode"`
}

var pipeline = make(chan IngestRequest, 10000)
var ConnStr = os.Getenv("CONNSTR")

func emptyStrToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intToNil(i int64) *int64 {
	if i == 0 {
		return nil
	}

	return &i
}

func handleRequests() {

	db, err := sql.Open("postgres", ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = MigrateDb(db)

	if err != nil {
		log.Fatal(err)
	}
	err = LoadIp2CountryDb(filepath.Join(Root, "dbip-country-lite.csv"))

	if err != nil {
		log.Fatal(err)
	}

	for {
		request := <-pipeline

		values, err := url.ParseQuery(request.Query)

		if err != nil {
			log.Error(err)
			continue
		}

		qv := make(map[string]string)
		var queryJson *[]byte

		if len(values) > 0 {
			for key := range values {
				qv[key] = values.Get(key)
			}

			qj, err := json.Marshal(qv)

			if err != nil {
				log.Errorf("Failed to serialize json: %s", err)
				continue
			}

			queryJson = &qj
		}

		country := GetCountry(request.ClientIp[0])

		if len(request.VisitorId) == 0 {
			h := sha256.New()
			h.Write([]byte(request.ClientIp[0] + request.ClientUserAgent))
			request.VisitorId = string(h.Sum(nil)[:])
		}

		// insert to events
		_, err = db.Exec("insert into public.events (\"timestamp\", \"domain\", event_name, duration, user_agent, referrer, path, visitor_id, query_params, country, status_code) values (NOW(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10);",
			strings.ToLower(strings.TrimSpace(request.Domain)), request.EventName, intToNil(request.Duration), request.ClientUserAgent, emptyStrToNil(request.Referrer), request.Path, request.VisitorId, queryJson, country, request.StatusCode)

		if err != nil {
			log.Errorf("Failed to insert event row: %s", err)
		}

		// insert into monthly traffic
		_, err = db.Exec("insert into public.monthly_traffic (timestamp, domain, duration, user_agent, referrer, path, query_params, country, status_code, ip, ips) values (NOW(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
			strings.ToLower(strings.TrimSpace(request.Domain)), intToNil(request.Duration), request.ClientUserAgent, emptyStrToNil(request.Referrer), request.Path, queryJson, country, request.StatusCode, request.ClientIp[0], pq.Array(request.ClientIp[1:]))

		if err != nil {
			log.Errorf("Failed to insert traffic row: %s", err)
		}
	}
}

func handleIngest(ctx iris.Context) {
	if ctx.GetHeader("Content-Type") != "application/json" {
		ctx.StatusCode(iris.StatusUnsupportedMediaType)
		return
	}

	var ingestBody IngestRequest
	err := ctx.ReadJSON(&ingestBody)

	if err != nil {
		ctx.StopWithError(400, err)
		return
	}

	if len(ingestBody.ClientIp) == 0 {
		ctx.StopWithError(iris.StatusBadRequest, fmt.Errorf("client ip is required"))
		return
	}

	if ingestBody.Domain == "" {
		ctx.StopWithError(iris.StatusBadRequest, fmt.Errorf("domain is required"))
		return
	}

	if ingestBody.Path == "" {
		ctx.StopWithError(iris.StatusBadRequest, fmt.Errorf("path is required"))
		return
	}

	if ingestBody.EventName == "" {
		ctx.StopWithError(iris.StatusBadRequest, fmt.Errorf("event name is required"))
		return
	}

	pipeline <- ingestBody

	ctx.StatusCode(iris.StatusOK)
}

func renderView(ctx iris.Context, name string, data map[string]interface{}) {
	if err := ctx.View(name, data); err != nil {
		log.WithFields(log.Fields{
			"error": fmt.Errorf("%w", err),
		}).Error("Failed to migrate database")
		_, _ = ctx.HTML("<h3 class=\"error\">%s</h3>", err.Error())
		return
	}
}

func handleStatsRequest(ctx iris.Context) {

	var start *time.Time
	var end *time.Time

	if ctx.URLParamExists("start") {
		t, err := time.Parse("2006-01-02", ctx.URLParam("start"))

		if err != nil {
			ctx.StopWithError(400, err)
			return
		}

		start = &t
	}

	if ctx.URLParamExists("end") {
		t, err := time.Parse("2006-01-02", ctx.URLParam("end"))

		if err != nil {
			ctx.StopWithError(400, err)
			return
		}

		end = &t
	}

	if start != nil && end != nil {
		if end.Before(*start) {
			ctx.StopWithError(400, fmt.Errorf("start time must be before end time"))
			return
		}
	}

	stats, err := GetStats("kilohearts.com", start, end)

	if err != nil {
		ctx.StopWithError(500, err)
		return
	}

	ctx.JSON(stats)
}

func main() {
	log.SetLevel(log.DebugLevel)
	app := iris.New()
	app.Logger().SetLevel("debug")

	tmpl := iris.Jet("./views", ".jet").Reload(true)
	app.RegisterView(tmpl)
	app.HandleDir("/public", iris.Dir("./public"))
	app.Post("/ingest", handleIngest)
	app.Get("/", func(ctx iris.Context) {
		renderView(ctx, "home", iris.Map{
			"StartDate": time.Now().Add(time.Hour * -24).Format("2006-01-02"),
			"EndDate":   time.Now().Format("2006-01-02"),
		})
	})
	app.Get("/stats", handleStatsRequest)

	go handleRequests()

	_ = app.Listen(":3100")
}
