package model

//TODO loose transaction levels wherever possible

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"sort"
	"time"
	"vkane.cz/tinyquiz/pkg/codeGenerator"
	"vkane.cz/tinyquiz/pkg/gameCreator"
	"vkane.cz/tinyquiz/pkg/model/ent"
	"vkane.cz/tinyquiz/pkg/model/ent/answer"
	"vkane.cz/tinyquiz/pkg/model/ent/askedquestion"
	"vkane.cz/tinyquiz/pkg/model/ent/choice"
	"vkane.cz/tinyquiz/pkg/model/ent/game"
	"vkane.cz/tinyquiz/pkg/model/ent/player"
	"vkane.cz/tinyquiz/pkg/model/ent/question"
	"vkane.cz/tinyquiz/pkg/model/ent/session"
	"vkane.cz/tinyquiz/pkg/rtcomm"
)

var NoSuchEntity = errors.New("no such entity found")
var ConstraintViolation = errors.New("constraint violation")

type Model struct {
	c *ent.Client
}

func NewModel(c *ent.Client) *Model {
	return &Model{c: c}
}

type Stats struct {
	Games    uint64
	Players  uint64
	Sessions uint64
}

func (m *Model) GetStats(c context.Context) (Stats, error) {
	var s Stats
	if games, err := m.c.Game.Query().Count(c); err == nil {
		s.Games = uint64(games)
	} else {
		return s, err
	}

	if players, err := m.c.Player.Query().Count(c); err == nil {
		s.Players = uint64(players)
	} else {
		return s, err
	}

	if sessions, err := m.c.Session.Query().Count(c); err == nil {
		s.Sessions = uint64(sessions)
	} else {
		return s, err
	}
	return s, nil
}

// returns the player's UUID if error is nil
// err = NoSuchEntity if the sessionCode is incorrect
func (m *Model) RegisterPlayer(playerName string, sessionCode string, now time.Time, c context.Context) (*ent.Player, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	if s, err := tx.Session.Query().Where(session.CodeEqualFold(sessionCode)).Only(c); ent.IsNotFound(err) {
		return nil, NoSuchEntity
	} else if err != nil {
		return nil, err
	} else {
		if p, err := tx.Player.Create().SetID(uuid.New()).SetJoined(now).SetName(playerName).SetSession(s).Save(c); err == nil {
			return p, nil
		} else if ent.IsConstraintError(err) {
			return nil, ConstraintViolation
		} else {
			return nil, err
		}
	}
}

func (m *Model) CreateSession(organiserName string, gameCode string, now time.Time, c context.Context) (*ent.Session, *ent.Player, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	if gameId, err := tx.Game.Query().Where(game.CodeEqualFold(gameCode)).OnlyID(c); err == nil {
		if incremental, err := m.getCodeIncremental(c); err == nil {
			if code, err := codeGenerator.GenerateRandomCode(incremental, codeRandomPartLength); err == nil {
				if s, err := tx.Session.Create().SetID(uuid.New()).SetCreated(now).SetCode(string(code)).SetGameID(gameId).Save(c); err == nil {
					if p, err := tx.Player.Create().SetID(uuid.New()).SetJoined(now).SetName(organiserName).SetSession(s).SetOrganiser(true).Save(c); err == nil {
						err := tx.Commit()
						return s, p, err
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	} else if ent.IsNotFound(err) {
		return nil, nil, NoSuchEntity
	} else {
		return nil, nil, err
	}
}

func (m *Model) GetPlayerWithSessionAndGame(uid uuid.UUID, c context.Context) (*ent.Player, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	if p, err := tx.Player.Query().Where(player.ID(uid)).WithSession(func(q *ent.SessionQuery) {
		q.WithGame()
	}).Only(c); err == nil {
		return p, nil
	} else if ent.IsNotFound(err) {
		return nil, NoSuchEntity
	} else {
		return nil, err
	}
}

func (m *Model) GetPlayersStateUpdate(sessionId uuid.UUID, c context.Context) (rtcomm.StateUpdate, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return rtcomm.StateUpdate{}, err
	}
	defer tx.Commit()

	if players, err := tx.Player.Query().Where(player.HasSessionWith(session.ID(sessionId))).Order(ent.Asc(player.FieldJoined)).All(c); err == nil {
		var su rtcomm.StateUpdate
		su.Players = make([]rtcomm.Player, 0, len(players))
		for i := 0; i < len(players); i++ {
			su.Players = append(su.Players, rtcomm.Player{
				Organiser: players[i].Organiser,
				Name:      players[i].Name,
			})
		}
		return su, nil
	} else {
		return rtcomm.StateUpdate{}, err
	}
}

func (m *Model) GetQuestionStateUpdate(sessionId uuid.UUID, now time.Time, c context.Context) (rtcomm.StateUpdate, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return rtcomm.StateUpdate{}, err
	}
	defer tx.Commit()

	if aq, err := tx.Question.Query().WithChoices().Where(question.HasAskedWith(askedquestion.HasSessionWith(session.ID(sessionId)))).QueryAsked().WithQuestion(func(q *ent.QuestionQuery) { q.WithChoices() }).Order(ent.Desc(askedquestion.FieldAsked)).First(c); err == nil {
		// either show the current question or hide the old one
		if aq.Ended == nil {
			var q = aq.Edges.Question
			var qu rtcomm.QuestionUpdate
			qu.Title = q.Title
			if ends := aq.Asked.Add(time.Duration(q.DefaultLength) * time.Millisecond); !now.Before(ends) {
				qu.RemainingTime = 0
			} else {
				qu.RemainingTime = uint64(ends.Sub(now).Round(time.Millisecond).Milliseconds())
			}
			qu.Answers = make([]rtcomm.Answer, 0, len(q.Edges.Choices))
			for i := 0; i < len(q.Edges.Choices); i++ {
				qu.Answers = append(qu.Answers, rtcomm.Answer{
					ID:    q.Edges.Choices[i].ID.String(),
					Title: q.Edges.Choices[i].Title,
				})
			}
			return rtcomm.StateUpdate{Question: &qu}, nil
		} else {
			return rtcomm.StateUpdate{Break: &rtcomm.BreakUpdate{}}, nil
		}
	} else if ent.IsNotFound(err) {
		// There is simply no current question, which is not an error
		return rtcomm.StateUpdate{}, nil
	} else {
		return rtcomm.StateUpdate{}, err
	}
}

//TODO reuse transaction
func (m *Model) GetFullStateUpdate(sessionId uuid.UUID, now time.Time, c context.Context) (rtcomm.StateUpdate, error) {
	su, err := m.GetPlayersStateUpdate(sessionId, c)
	if err != nil {
		return rtcomm.StateUpdate{}, err
	}
	if su2, err := m.GetQuestionStateUpdate(sessionId, now, c); err == nil {
		su.Question = su2.Question
	} else {
		return rtcomm.StateUpdate{}, err
	}
	return su, nil
}

var NoNextQuestion = errors.New("there is no next question") // TODO fill

// TODO retry on serialization failure
// TODO validate sessionId
func (m *Model) NextQuestion(sessionId uuid.UUID, now time.Time, c context.Context) error {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	// TODO rollback only if not yet committed
	defer tx.Rollback()

	if err := tx.Session.Update().Where(session.ID(sessionId)).Where(session.StartedIsNil()).SetStarted(now).Exec(c); err != nil {
		return err
	}

	var query = tx.Question.Query().Where(question.HasGameWith(game.HasSessionsWith(session.ID(sessionId)))).Order(ent.Asc(question.FieldOrder))

	if current, err := tx.AskedQuestion.Query().Where(askedquestion.HasSessionWith(session.ID(sessionId))).WithQuestion().Order(ent.Desc(askedquestion.FieldAsked)).First(c); err == nil {
		query.Where(question.OrderGT(current.Edges.Question.Order))
		if current.Ended == nil {
			if _, err := current.Update().SetEnded(now).Save(c); err == nil {
				return tx.Commit()
			} else {
				return err
			}
		}
	} else if !ent.IsNotFound(err) {
		return err
	}

	if next, err := query.First(c); err == nil {
		if _, err := tx.AskedQuestion.Create().SetID(uuid.New()).SetAsked(now).SetSessionID(sessionId).SetQuestion(next).Save(c); err != nil {
			return err
		}
	} else if ent.IsNotFound(err) {
		return NoNextQuestion
	} else {
		return err
	}

	tx.Commit()
	return nil
}

var QuestionClosed = errors.New("the deadline for answers to this question has passed")
var AlreadyAnswered = errors.New("the player has already answered the question")

func (m *Model) SaveAnswer(playerId uuid.UUID, choiceId uuid.UUID, now time.Time, c context.Context) (*ent.Answer, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// check whether the player could pick this choice
	if exists, err := tx.Choice.Query().Where(choice.HasQuestionWith(question.HasGameWith(game.HasSessionsWith(session.HasPlayersWith(player.ID(playerId))))), choice.ID(choiceId)).Exist(c); err == nil && !exists {
		return nil, NoSuchEntity
	} else if err != nil {
		return nil, err
	}

	var q *ent.Question
	// find the most recent question
	if q, err = tx.Player.Query().Where(player.ID(playerId)).QuerySession().QueryAskedQuestions().QueryQuestion().WithAsked(func(q *ent.AskedQuestionQuery) {
		q.Where(askedquestion.HasSessionWith(session.HasPlayersWith(player.ID(playerId))))
	}).Order(ent.Desc(question.FieldOrder)).First(c); ent.IsNotFound(err) {
		return nil, NoSuchEntity
	} else if err != nil {
		return nil, err
	}

	// check if the question is open
	// Asked[0] is guaranteed to exist thanks to the previous query
	if q.Edges.Asked[0].Ended != nil || q.Edges.Asked[0].Asked.Add(time.Duration(q.DefaultLength)*time.Millisecond).Before(now) {
		return nil, QuestionClosed
	}

	// check the player has not answered yet
	if exists, err := tx.Answer.Query().Where(answer.HasAnswererWith(player.ID(playerId)), answer.HasChoiceWith(choice.HasQuestionWith(question.ID(q.ID)))).Exist(c); err != nil {
		return nil, err
	} else if exists {
		return nil, AlreadyAnswered
	}

	if a, err := tx.Answer.Create().SetID(uuid.New()).SetAnswered(now).SetChoiceID(choiceId).SetAnswererID(playerId).Save(c); err == nil {
		tx.Commit()
		return a, nil
	} else {
		return nil, err
	}
}

type PlayerResult struct {
	Player  *ent.Player
	place   uint64
	correct int64
}

func (r PlayerResult) Points() int64 {
	return r.correct
}

func (r PlayerResult) Place() uint64 {
	return r.place
}

func (m *Model) GetResults(playerId uuid.UUID, c context.Context) ([]PlayerResult, *ent.Session, *ent.Player, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Commit()

	s, err := tx.Session.Query().WithGame().Where(session.HasPlayersWith(player.ID(playerId))).Only(c)
	if ent.IsNotFound(err) {
		return nil, nil, nil, NoSuchEntity
	} else if err != nil {
		return nil, nil, nil, err
	}

	p, err := tx.Player.Query().Where(player.ID(playerId)).Only(c)
	if ent.IsNotFound(err) {
		return nil, nil, nil, NoSuchEntity
	} else if err != nil {
		return nil, nil, nil, err
	}

	if players, err := tx.Player.Query().Where(player.HasSessionWith(session.ID(s.ID))).Where(player.Organiser(false)).Order(ent.Asc(player.FieldName)).WithAnswers(func(q *ent.AnswerQuery) { q.WithChoice() }).All(c); err == nil {
		var results = make([]PlayerResult, 0, len(players))
		for _, p := range players {
			var res PlayerResult
			res.Player = p
			for _, a := range p.Edges.Answers {
				if a.Edges.Choice.Correct {
					res.correct++
				}
			}
			results = append(results, res)
		}
		sort.SliceStable(results, func(i, j int) bool { return results[i].correct > results[j].correct }) // sort in reverse
		if len(results) > 0 {
			results[0].place = 1
		}
		var place uint64 = 2
		for i := 1; i < len(results); i++ {
			if results[i].Points() == results[i-1].Points() {
				results[i].place = results[i-1].place
			} else {
				results[i].place = place
			}
			place++
		}
		return results, s, p, nil
	} else {
		return nil, nil, nil, err
	}
}

func (m *Model) CreateGame(game gameCreator.Game, name string, author string, c context.Context) (*ent.Game, error) {
	tx, err := m.c.BeginTx(c, &sql.TxOptions{
		Isolation: sql.LevelReadUncommitted,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var code []byte
	if incremental, err := m.getCodeIncremental(c); err == nil {
		if c, err := codeGenerator.GenerateRandomCode(incremental, codeRandomPartLength); err == nil {
			code = c
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	g, err := tx.Game.Create().SetID(uuid.New()).SetCreated(time.Now()).SetName(name).SetAuthor(author).SetCode(string(code)).Save(c)
	if err != nil {
		return nil, err
	}

	var questions = make([]*ent.QuestionCreate, 0, len(game.Questions))
	var questionIds = make([]uuid.UUID, 0, len(game.Questions))
	var choicesCount uint
	for i, q := range game.Questions {
		var id = uuid.New()
		var questionCreate = tx.Question.Create().SetID(id).SetGame(g).SetDefaultLength(q.Length).SetOrder(i + 1).SetTitle(q.Title)
		questions = append(questions, questionCreate)
		questionIds = append(questionIds, id)
		choicesCount += uint(len(q.Choices))
	}
	if _, err := tx.Question.CreateBulk(questions...).Save(c); err != nil {
		return nil, err
	}

	var choices = make([]*ent.ChoiceCreate, 0, choicesCount)
	for i, q := range game.Questions {
		for _, c := range q.Choices {
			var choiceCreate = tx.Choice.Create().SetID(uuid.New()).SetTitle(c.Title).SetCorrect(c.Correct).SetQuestionID(questionIds[i])
			choices = append(choices, choiceCreate)
		}
	}
	if _, err := tx.Choice.CreateBulk(choices...).Save(c); err != nil {
		return nil, err
	}
	return g, tx.Commit()
}

func (m *Model) GetGameWithQuestionsAndChoices(gameId uuid.UUID, c context.Context) (*ent.Game, error) {
	if game, err := m.c.Game.Query().Where(game.ID(gameId)).WithQuestions(func(q *ent.QuestionQuery) { q.WithChoices() }).Only(c); err == nil {
		return game, err
	} else if ent.IsNotFound(err) {
		return nil, NoSuchEntity
	} else {
		return nil, err
	}
}

const codeRandomPartLength uint8 = 3

func (m *Model) getCodeIncremental(c context.Context) (uint64, error) {
	if c, err := m.c.CodesSequence.Create().Save(c); err == nil {
		return uint64(c.ID), nil
	} else {
		return 0, nil
	}
}
