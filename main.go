package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type IngestRequest struct {
	Domain          string `json:"domain"`
	Path            string `json:"path"`
	Query           string `json:"query"`
	EventName       string `json:"eventName"`
	VisitorId       string `json:"visitorId"`
	Referrer        string `json:"referrer"`
	ClientIp        string `json:"clientIp"`
	ClientUserAgent string `json:"clientUserAgent"`
	Duration        int64  `json:"duration"`
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

		country := GetCountry(request.ClientIp)

		_, err = db.Exec("\ninsert into public.traffic (\"timestamp\", \"domain\", event_name, duration, user_agent, referrer, path, visitor_id, query_params, country) values (NOW(), $1, $2, $3, $4, $5, $6, $7, $8, $9);",
			strings.ToLower(strings.TrimSpace(request.Domain)), request.EventName, intToNil(request.Duration), request.ClientUserAgent, emptyStrToNil(request.Referrer), request.Path, request.VisitorId, queryJson, country)

		if err != nil {
			log.Fatalf("Failed to insert row: %s", err)
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

	if ingestBody.Domain == "" {
		ctx.StopWithError(400, fmt.Errorf("domain is required"))
		return
	}

	if ingestBody.Path == "" {
		ctx.StopWithError(400, fmt.Errorf("path is required"))
		return
	}

	if ingestBody.EventName == "" {
		ctx.StopWithError(400, fmt.Errorf("event name is required"))
		return
	}

	if ingestBody.VisitorId == "" {
		ctx.StopWithError(400, fmt.Errorf("visitor id is required"))
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

	stats, err := GetStats("kilohearts.com", nil, nil)

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
	app.Get("/", func(ctx iris.Context) { renderView(ctx, "home", iris.Map{}) })
	app.Get("/stats", handleStatsRequest)

	go handleRequests()

	_ = app.Listen("localhost:3100")
}
