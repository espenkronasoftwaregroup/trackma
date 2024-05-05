package main

import (
	"database/sql"
	"github.com/kataras/iris/v12"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"io"
)

type IngestRequest struct {
	OriginIp    string
	Path        string
	QueryParams map[string]string
	Body        string
	Referrer    string
	UserAgent   string
	GroupId     string
}

var pipeline = make(chan IngestRequest, 10000)
var ips *[]IpRangeCountry

func handleRequests() {

	db, err := sql.Open("traffic", "postgresql://<username>:<password>@<database_ip>/todos?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = MigrateDb(db)

	if err != nil {
		log.Fatal(err)
	}
	//https://pkg.go.dev/github.com/mostafa-asg/ip2country#section-readme check
	for {
		request := <-pipeline

		if err != nil {
			log.Errorf("Faild to do geo ip lookup: %s", err.Error())
			continue
		}

		db.Exec("INSERT INTO traffic VALUES()")
	}
}

func handleIngest(ctx iris.Context) {
	if ctx.GetHeader("Content-Type") != "application/json" {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(ctx.Request().Body)

	if err != nil {
		log.Errorf("Failed to parse request json: %s", err.Error())
		ctx.StopWithError(400, err)
		return
	}

	clientIp := ctx.GetHeader("X-Forwarded-For")

	if clientIp == "" {
		clientIp = ctx.RemoteAddr()
	}

	var request = IngestRequest{
		Body:        string(body[:]),
		OriginIp:    clientIp,
		Referrer:    ctx.GetHeader("Referrer"),
		UserAgent:   ctx.GetHeader("User-Agent"),
		Path:        ctx.Path(),
		QueryParams: ctx.URLParams(),
	}

	pipeline <- request

	ctx.StatusCode(iris.StatusOK)
}

func handleMainGet(ctx iris.Context) {
	ctx.HTML("")
}

func main() {
	log.SetLevel(log.DebugLevel)
	app := iris.New()
	app.Logger().SetLevel("debug")

	i, err := ReadDbIpCsv()

	if err != nil {
		log.Fatal(err.Error())
	}

	ips = i

	tmpl := iris.Jet("./views", ".jet").Reload(true)
	app.RegisterView(tmpl)
	app.HandleDir("/public", iris.Dir("./public"))
	app.Post("/ingest", handleIngest)
	app.Get("/", handleMainGet)

	go handleRequests()

	_ = app.Listen(":3000")
}
