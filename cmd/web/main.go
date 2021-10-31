package main

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"vkane.cz/tinyquiz/pkg/model"
	"vkane.cz/tinyquiz/pkg/model/ent"
	"vkane.cz/tinyquiz/pkg/model/ent/migrate"
	rtcomm "vkane.cz/tinyquiz/pkg/rtcomm"
	"vkane.cz/tinyquiz/ui"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	model         *model.Model
	rtClients     *rtcomm.Clients
}

type templateData struct {
}

func setDefaultTemplateData(td *templateData) {
}

func main() {
	var addr string
	var socket bool
	if env, ok := os.LookupEnv("TINYQUIZ_LISTEN"); ok {
		addr = env
		const unixPrefix = "unix:"
		if strings.HasPrefix(addr, unixPrefix) {
			socket = true
			addr = strings.TrimPrefix(addr, unixPrefix)
		}
	} else {
		addr = "[::1]:8080"
	}

	var pgConnectionUri url.URL
	pgConnectionUri.Scheme = "postgresql"
	pgConnectionUri.Path = "/"
	pgQuery := pgConnectionUri.Query()
	if env, ok := os.LookupEnv("TINYQUIZ_PG_HOST"); ok {
		pgQuery.Set("host", env)
	} else {
		pgQuery.Set("host", "127.0.0.1")
	}
	if env, ok := os.LookupEnv("TINYQUIZ_PG_DBNAME"); ok {
		pgQuery.Set("dbname", env)
	} else {
		pgQuery.Set("dbname", "tinyquiz")
	}
	if env, ok := os.LookupEnv("TINYQUIZ_PG_USER"); ok {
		pgQuery.Set("user", env)
	}
	if env, ok := os.LookupEnv("TINYQUIZ_PG_PASSWORD"); ok {
		pgQuery.Set("password", env)
	}
	if _, ok := os.LookupEnv("TINYQUIZ_PG_ENABLESSL"); ok {
		pgQuery.Set("sslmode", "verify-full")
	} else {
		pgQuery.Set("sslmode", "disable")
	}
	pgConnectionUri.RawQuery = pgQuery.Encode()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var app = application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		rtClients: rtcomm.NewClients(),
	}

	if tc, err := newTemplateCache(); err == nil {
		app.templateCache = tc
	} else {
		errorLog.Fatal(err)
	}

	if c, err := ent.Open("postgres", pgConnectionUri.String()); err == nil {
		if err := c.Schema.Create(context.Background(), migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
			errorLog.Fatal(err)
		}
		app.model = model.NewModel(c)
	} else {
		errorLog.Fatal(err)
	}

	//TODO remove debug print
	go func() {
		for range time.Tick(2 * time.Second) {
			//sessions, clients := app.rtClients.Count()
			//fmt.Printf("There are %d sessions with total of %d clients\n", sessions, clients)
		}
	}()

	mux := httprouter.New()
	mux.GET("/", app.homeSuccess)
	mux.POST("/play/:code", app.play)
	mux.POST("/session", app.createSession)
	mux.GET("/game/:playerUid", app.game)
	mux.POST("/game/:playerUid/rpc/next", app.nextQuestion)
	mux.POST("/game/:playerUid/answers/:choiceUid", app.answer)
	mux.GET("/results/:playerUid", app.resultsGeneral)
	mux.GET("/template", app.downloadTemplate)
	mux.POST("/game", app.createGame)
	mux.GET("/quiz/:gameUid", app.showGame)
	mux.GET("/help", app.help)

	mux.GET("/ws/:playerUid", app.processWebSocket)

	if static, err := fs.Sub(ui.StaticFiles, "static"); err == nil {
		mux.ServeFiles("/static/*filepath", http.FS(static))
	} else {
		errorLog.Fatal(err)
	}

	var srv = &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}
	log.Printf("Starting server on %s\n", addr)
	if socket {
		if listener, err := net.Listen("unix", addr); err == nil {
			log.Fatal(srv.Serve(listener))
		} else {
			errorLog.Fatal(err)
		}
	} else {
		errorLog.Fatal(srv.ListenAndServe())
	}
}
