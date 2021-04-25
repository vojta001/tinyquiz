package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
	"vkane.cz/tinyquiz/pkg/model"
	"vkane.cz/tinyquiz/pkg/model/ent"
	"vkane.cz/tinyquiz/pkg/rtcomm"
	"vkane.cz/tinyquiz/ui"
)

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	templates, err := fs.Sub(ui.HTMLTemplates, "html")
	if err != nil {
		return nil, err
	}

	pages, err := fs.Glob(templates, "*.page.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFS(templates, page)
		if err != nil {
			return nil, err
		}

		tsL, err := ts.ParseFS(templates, "*.layout.tmpl.html")
		if err != nil && !strings.HasPrefix(err.Error(), "template: pattern matches no files") {
			return nil, err
		} else if err == nil {
			ts = tsL
		}

		tsP, err := ts.ParseFS(templates, "*.partial.tmpl.html")
		if err != nil && !strings.HasPrefix(err.Error(), "template: pattern matches no files") {
			return nil, err
		} else if err == nil {
			ts = tsP
		}

		cache[name] = ts
	}

	return cache, nil
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td interface{}) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	err := ts.Execute(w, td)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	app.errorLog.Output(2, string(debug.Stack()))

	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1, // we will not be reading from the socket
	EnableCompression: true,
}
const suBufferSize = 100
const writeDeadline = time.Second * 10
const socketGCPeriod = writeDeadline

// TODO utilize request context
func (app *application) processWebSocket(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var playerUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("playerUid")); err == nil {
		playerUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var player *ent.Player
	if p, err := app.model.GetPlayerWithSessionAndGame(playerUid, r.Context()); err == nil {
		player = p
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}

	if c, err := upgrader.Upgrade(w, r, nil); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	} else {
		defer c.Close()
		if tcp, ok := c.UnderlyingConn().(*net.TCPConn); ok {
			tcp.SetKeepAlivePeriod(writeDeadline)
			tcp.SetKeepAlive(true)
		} else {
			app.infoLog.Printf("could not set keepalive for %s\n", playerUid)
		}
		var ch = make(chan rtcomm.StateUpdate, suBufferSize)
		app.rtClients.AddClient(player.Edges.Session.ID, ch)
		defer app.rtClients.RemoveClient(player.Edges.Session.ID, ch)
		if su, err := app.model.GetFullStateUpdate(player.Edges.Session.ID, time.Now(), r.Context()); err == nil {
			select {
			case ch <- su:
				break
			default:
				app.infoLog.Printf("could not send initial StateUpdate to %s\n", playerUid.String())
			}
		} else {
			app.errorLog.Printf("failed getting initial StateUpdate for %s with %v\n", playerUid.String(), err)
		}
		var gcTicker = time.Tick(socketGCPeriod)
		var devnull [0]byte
	loop:
		for {
			select {
			case <- gcTicker:
				c.UnderlyingConn().SetWriteDeadline(time.Time{})
				if _, err := c.UnderlyingConn().Write(devnull[:]); err != nil {
					app.infoLog.Printf("closing broken (%v) connection of %s\n", err, playerUid.String())
					break loop
				}
			case su := <- ch:
				if err := c.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
					app.errorLog.Printf("setting write deadline for %s failed with %v\n", playerUid.String(), err)
					break loop
				}
				if err := c.WriteJSON(su); errors.Is(err, io.EOF) {
					break loop
				} else if err != nil {
					app.infoLog.Printf("sending message for %s failed with %v\n", playerUid.String(), err)
					break loop
				}
			}
		}
	}
}
