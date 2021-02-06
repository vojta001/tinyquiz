package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"strings"
	"vkane.cz/tinyquiz/pkg/model"
	"vkane.cz/tinyquiz/pkg/model/ent"
)

func (app *application) home(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type homeData struct {
		Stats model.Stats
		templateData
	}
	td := &homeData{}
	setDefaultTemplateData(&td.templateData)

	if stats, err := app.model.GetStats(r.Context()); err == nil {
		td.Stats = stats
	} else {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl.html", td)
}

func (app *application) play(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var code = strings.TrimSpace(params.ByName("code"))
	var player = strings.ToLower(strings.TrimSpace(r.PostForm.Get("player")))

	if len(player) < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if player, err := app.model.RegisterPlayer(player, code, r.Context()); err == nil {
		if session, err := player.Unwrap().QuerySession().Only(r.Context()); err == nil {
			if su, err := app.model.GetPlayersStateUpdate(session.ID, r.Context()); err == nil {
				app.rtClients.SendToAll(session.ID, su)
			} else {
				app.serverError(w, err)
				return
			}
		} else {
			app.serverError(w, err)
			return
		}
		http.Redirect(w, r, "/game/"+url.PathEscape(player.ID.String()), http.StatusSeeOther)
		return
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else if errors.Is(err, model.ConstraintViolation) {
		app.clientError(w, http.StatusForbidden)
		return
	} else {
		app.serverError(w, err)
		return
	}

}

func (app *application) game(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var playerUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("playerUid")); err == nil {
		playerUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if player, err := app.model.GetPlayerWithSessionAndGame(playerUid, r.Context()); err == nil {
		type lobbyData struct {
			P *ent.Player
			templateData
		}
		td := &lobbyData{}
		setDefaultTemplateData(&td.templateData)
		td.P = player

		app.render(w, r, "game.page.tmpl.html", td)
	} else {
		app.clientError(w, http.StatusNotFound)
		return
	}
}

func (app *application) nextQuestion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var playerUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("playerUid")); err == nil {
		playerUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if player, err := app.model.GetPlayerWithSessionAndGame(playerUid, r.Context()); err == nil {
		var sessionId = player.Edges.Session.ID
		if err := app.model.NextQuestion(sessionId, r.Context()); err == nil {
			if su, err := app.model.GetQuestionStateUpdate(sessionId, r.Context()); err == nil {
				app.rtClients.SendToAll(sessionId, su)
			} else {
				app.serverError(w, err)
				return
			}
		} else {
			app.serverError(w, err)
			return
		}
	} else if ent.IsNotFound(err) {
		app.infoLog.Println(playerUid)
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}
}
