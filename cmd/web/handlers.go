package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"strings"
	"time"
	"vkane.cz/tinyquiz/pkg/gameCreator"
	"vkane.cz/tinyquiz/pkg/model"
	"vkane.cz/tinyquiz/pkg/model/ent"
	"vkane.cz/tinyquiz/pkg/rtcomm"
)

func (app *application) homeSuccess(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	app.home(w, r, homeForm{}, http.StatusOK)
}

type homeForm struct {
	Join struct {
		Code   string
		Name   string
		Errors []string
	}
	NewSession struct {
		Code   string
		Name   string
		Errors []string
	}
	NewGame struct {
		Title  string
		Name   string
		Errors []string
	}
}

func (app *application) home(w http.ResponseWriter, r *http.Request, formData homeForm, status int) {
	type homeData struct {
		Stats model.Stats
		Form  homeForm
		templateData
	}
	td := &homeData{}
	td.Form = formData
	setDefaultTemplateData(&td.templateData)

	if stats, err := app.model.GetStats(r.Context()); err == nil {
		td.Stats = stats
	} else {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(status)
	app.render(w, r, "home.page.tmpl.html", td)
}

func (app *application) help(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	app.render(w, r, "help.page.tmpl.html", nil)
}

func (app *application) play(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var code = strings.TrimSpace(params.ByName("code"))
	var player = strings.ToLower(strings.TrimSpace(r.PostForm.Get("player")))
	var form homeForm
	form.Join.Code = code
	form.Join.Name = player

	if len(player) < 1 {
		form.Join.Errors = []string{"Zadejte jméno hráče"}
		app.home(w, r, form, http.StatusBadRequest)
		return
	}

	if player, err := app.model.RegisterPlayer(player, code, time.Now(), r.Context()); err == nil {
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
		form.Join.Errors = []string{"Hra s tímto kódem nebyla nalezena"}
		app.home(w, r, form, http.StatusNotFound)
		return
	} else if errors.Is(err, model.ConstraintViolation) {
		form.Join.Errors = []string{"Hráč s tímto jménem již existuje"}
		app.home(w, r, form, http.StatusForbidden)
		return
	} else {
		app.serverError(w, err)
		return
	}

}

func (app *application) createSession(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var code = strings.TrimSpace(r.PostForm.Get("code"))
	var player = strings.ToLower(strings.TrimSpace(r.PostForm.Get("organiser")))
	var form homeForm
	form.NewSession.Code = code
	form.NewSession.Name = player

	if len(player) < 1 {
		form.NewSession.Errors = []string{"Zadejte jméno organizátora"}
		app.home(w, r, form, http.StatusBadRequest)
		return
	}

	if s, p, err := app.model.CreateSession(player, code, time.Now(), r.Context()); err == nil {
		if su, err := app.model.GetPlayersStateUpdate(s.ID, r.Context()); err == nil {
			app.rtClients.SendToAll(s.ID, su)
			http.Redirect(w, r, "/game/"+url.PathEscape(p.ID.String()), http.StatusSeeOther)
			return
		} else {
			app.serverError(w, err)
			return
		}
	} else if errors.Is(err, model.NoSuchEntity) {
		form.NewSession.Errors = []string{"Hra s tímto kódem nebyla nalezena"}
		app.home(w, r, form, http.StatusNotFound)
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
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
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
		var now = time.Now()
		if err := app.model.NextQuestion(sessionId, now, r.Context()); err == nil {
			if su, err := app.model.GetQuestionStateUpdate(sessionId, now, r.Context()); err == nil {
				app.rtClients.SendToAll(sessionId, su)
				w.WriteHeader(http.StatusNoContent)
				return
			} else {
				app.serverError(w, err)
				return
			}
		} else if errors.Is(err, model.NoNextQuestion) {
			app.rtClients.SendToAll(sessionId, rtcomm.StateUpdate{Results: true})
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			app.serverError(w, err)
			return
		}
	} else if ent.IsNotFound(err) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}
}

func (app *application) answer(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var playerUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("playerUid")); err == nil {
		playerUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var choiceUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("choiceUid")); err == nil {
		choiceUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := app.model.SaveAnswer(playerUid, choiceUid, time.Now(), r.Context()); err == nil {
		// TODO notify organisers
		w.WriteHeader(http.StatusCreated) // TODO or StatusNoContent?
		return
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}
}

func (app *application) resultsGeneral(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type resultsData struct {
		Results []model.PlayerResult
		Session *ent.Session
		Player  *ent.Player
		templateData
	}
	td := &resultsData{}
	setDefaultTemplateData(&td.templateData)

	var playerUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("playerUid")); err == nil {
		playerUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if results, session, player, err := app.model.GetResults(playerUid, r.Context()); err == nil {
		td.Results = results
		td.Session = session
		td.Player = player
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "results.page.tmpl.html", td)
}

func (app *application) downloadTemplate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8; header=absent")
	w.Header().Set("Content-Disposition", "attachment; filename=tinyquiz_template.csv")
	w.Header().Set("Cache-Control", "max-age=21600" /* 6 hours */)
	if err := gameCreator.CreateTemplate(w, 10, 4); err != nil {
		app.errorLog.Printf("creating template: %v", err)
		return
	}
}

func (app *application) createGame(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, 1000000)
	// shall never write to temp files thanks to MaxBytesReader
	if err := r.ParseMultipartForm(2000000); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var name = r.PostFormValue("name")
	var author = r.PostFormValue("author")
	var form homeForm
	form.NewGame.Title = name
	form.NewGame.Name = author
	file, _, err := r.FormFile("game")
	if err != nil {
		form.NewGame.Errors = []string{"Nahrajte soubor s otázkami"}
		app.home(w, r, form, http.StatusBadRequest)
		return
	}

	if parsedGame, err := gameCreator.Parse(file, 500, 100); err == nil {
		if game, err := app.model.CreateGame(parsedGame, name, author, r.Context()); err == nil {
			http.Redirect(w, r, "/quiz/"+url.PathEscape(game.ID.String()), http.StatusSeeOther)
			return
		} else {
			app.serverError(w, err)
			return
		}
	} else {
		form.NewGame.Errors = []string{"Soubor s otázkami není v pořádku"}
		app.home(w, r, form, http.StatusBadRequest)
		return
	}
}

func (app *application) showGame(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type gameData struct {
		Game *ent.Game
		templateData
	}
	td := &gameData{}
	setDefaultTemplateData(&td.templateData)

	var gameUid uuid.UUID
	if uid, err := uuid.Parse(params.ByName("gameUid")); err == nil {
		gameUid = uid
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if game, err := app.model.GetGameWithQuestionsAndChoices(gameUid, r.Context()); err == nil {
		td.Game = game
		w.Header().Set("Cache-Control", "max-age=3600" /* 1 hour */)
		app.render(w, r, "game-overview.page.tmpl.html", td)
		return
	} else if errors.Is(err, model.NoSuchEntity) {
		app.clientError(w, http.StatusNotFound)
		return
	} else {
		app.serverError(w, err)
		return
	}
}
