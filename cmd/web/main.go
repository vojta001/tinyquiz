package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"vkane.cz/tinyquiz/pkg/model"
	"vkane.cz/tinyquiz/pkg/model/ent"
	"vkane.cz/tinyquiz/pkg/model/ent/migrate"
	rtcomm "vkane.cz/tinyquiz/pkg/rtcomm"

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
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var app = application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		rtClients: rtcomm.NewClients(),
	}

	if tc, err := newTemplateCache("./ui/html/"); err == nil {
		app.templateCache = tc
	} else {
		errorLog.Fatal(err)
	}

	if c, err := ent.Open("postgres", "host='127.0.0.1' sslmode=disable dbname=tinyquiz"); err == nil {
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
	mux.GET("/", app.home)
	mux.POST("/play/:code", app.play)
	mux.POST("/session", app.createSession)
	mux.GET("/game/:playerUid", app.game)
	mux.POST("/game/:playerUid/rpc/next", app.nextQuestion)
	mux.POST("/game/:playerUid/answers/:choiceUid", app.answer)
	mux.GET("/results/:playerUid", app.resultsGeneral)

	mux.GET("/ws/:playerUid", app.processWebSocket)

	mux.ServeFiles("/static/*filepath", http.Dir("./ui/static/"))

	var srv = &http.Server{
		Addr:     "127.0.0.1:8080",
		ErrorLog: errorLog,
		Handler:  mux,
	}
	log.Println("Starting server on :8080")
	err := srv.ListenAndServe()
	log.Fatal(err)
}
