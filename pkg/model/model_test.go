package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"net/url"
	"testing"
	"time"
	"vkane.cz/tinyquiz/pkg/model/ent"
)

func newTestDb(t *testing.T) *ent.Client {
	c, err := ent.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=private&_fk=1", url.PathEscape(t.Name())))
	ent.NewClient()
	if err != nil {
		t.Fatalf("Could not create temporary database: %v", err)
	}
	if err = c.Schema.Create(context.Background()); err != nil {
		t.Fatalf("Could not initialize schema in temporary database: %v", err)
	}
	t.Cleanup(func() {
		c.Close()
	})
	return c
}

func newTestModel(t *testing.T) *Model {
	return NewModel(newTestDb(t))
}

func newTestModelWithData(t *testing.T) *Model {
	m := newTestModel(t)

	c := context.Background()

	tx, err := m.c.BeginTx(c, nil)
	if err != nil {
		t.Fatalf("Error during default data insertion: %v", err)
	}
	defer tx.Rollback()

	var gamesC = []*ent.GameCreate{
		tx.Game.Create().SetID(uuid.MustParse("cab48de7-bba3-4873-9335-eec4aaaae1e9")).SetName("5th grade knowledge test").SetCreated(time.Unix(1613387448, 0)).SetAuthor("Adam Smith PhD."),
	}

	games := tx.Game.CreateBulk(gamesC...).SaveX(c)

	var questionsC = []*ent.QuestionCreate{
		tx.Question.Create().SetID(uuid.MustParse("65b848a8-7d0e-4b16-96aa-c6b89bda6657")).SetTitle("The WWII ended in:").SetOrder(1).SetDefaultLength(30000).SetGame(games[0]),
		tx.Question.Create().SetID(uuid.MustParse("adb9b601-9ae7-4d91-8998-968d9848eeb4")).SetTitle("What is the capital of the USA?").SetOrder(2).SetDefaultLength(30000).SetGame(games[0]),
	}

	questions := tx.Question.CreateBulk(questionsC...).SaveX(c)

	var choicesC = []*ent.ChoiceCreate{
		tx.Choice.Create().SetID(uuid.MustParse("7be00601-d316-46ef-842d-d7b25235905f")).SetTitle("1945").SetCorrect(true).SetQuestion(questions[0]),
		tx.Choice.Create().SetID(uuid.MustParse("9bd328e9-7a6f-4c39-9d91-7302a5916eeb")).SetTitle("1944").SetCorrect(false).SetQuestion(questions[0]),
		tx.Choice.Create().SetID(uuid.MustParse("b88b7f4e-1b17-49ea-8e90-cf42ae4e0f09")).SetTitle("1845").SetCorrect(false).SetQuestion(questions[0]),
		tx.Choice.Create().SetID(uuid.MustParse("5155b997-eb2c-4cd0-a067-2bb01379730f")).SetTitle("1549").SetCorrect(false).SetQuestion(questions[0]),

		tx.Choice.Create().SetID(uuid.MustParse("01819d92-fb00-4543-827e-b44f8ba17854")).SetTitle("New York").SetCorrect(false).SetQuestion(questions[1]),
		tx.Choice.Create().SetID(uuid.MustParse("d438e6de-a142-4cc5-9c0f-1bb3a37786c0")).SetTitle("Los Angeles").SetCorrect(false).SetQuestion(questions[1]),
		tx.Choice.Create().SetID(uuid.MustParse("77872cdd-db89-451d-87d3-0804e6f99e5e")).SetTitle("Washington DC").SetCorrect(true).SetQuestion(questions[1]),
		tx.Choice.Create().SetID(uuid.MustParse("4fd20819-525c-4172-8e4e-7f82585b6c23")).SetTitle("Berlin").SetCorrect(false).SetQuestion(questions[1]),
	}

	choices := tx.Choice.CreateBulk(choicesC...).SaveX(c)

	var sessionsC = []*ent.SessionCreate{
		tx.Session.Create().SetID(uuid.MustParse("b3d2f5b2-d5eb-4461-b352-622431a35b12")).SetCreated(time.Unix(1613387962, 0)).SetStarted(time.Unix(1613388071, 0)).SetCode("abcdef").SetGame(games[0]),
	}

	sessions := tx.Session.CreateBulk(sessionsC...).SaveX(c)

	var playersC = []*ent.PlayerCreate{
		tx.Player.Create().SetID(uuid.MustParse("fccc652f-e674-4c4f-9d45-6938090d3df1")).SetName("A. Smith").SetJoined(time.Unix(1613387963, 0)).SetOrganiser(true).SetSession(sessions[0]),
		tx.Player.Create().SetID(uuid.MustParse("15ca5cdb-d26a-42de-a9b3-29c8b3095296")).SetName("M. Black").SetJoined(time.Unix(1613387965, 0)).SetOrganiser(true).SetSession(sessions[0]),
		tx.Player.Create().SetID(uuid.MustParse("f8cd85a4-8b46-4145-abaf-df924a7719cf")).SetName("Bob").SetJoined(time.Unix(1613387969, 0)).SetOrganiser(false).SetSession(sessions[0]),
		tx.Player.Create().SetID(uuid.MustParse("321f3bb4-f789-49db-ad14-45299a4725a0")).SetName("Lisa ❤️").SetJoined(time.Unix(1613387975, 0)).SetOrganiser(false).SetSession(sessions[0]),
		tx.Player.Create().SetID(uuid.MustParse("cd0afe61-2c89-473f-9269-bbcb50016941")).SetName("Petr").SetJoined(time.Unix(1613387976, 0)).SetOrganiser(false).SetSession(sessions[0]),
	}

	players := tx.Player.CreateBulk(playersC...).SaveX(c)

	var askedQuestionsC = []*ent.AskedQuestionCreate{
		tx.AskedQuestion.Create().SetID(uuid.MustParse("72a1bb9c-67e7-4d59-80fa-80ce729629d3")).SetAsked(time.Unix(1613387996, 0)).SetQuestion(questions[0]).SetSession(sessions[0]).SetEnded(time.Unix(1613388001, 0)),
	}

	askedQuestions := tx.AskedQuestion.CreateBulk(askedQuestionsC...).SaveX(c)

	var answersC = []*ent.AnswerCreate{
		tx.Answer.Create().SetID(uuid.MustParse("387e626f-aed1-4bb3-953f-744763018178")).SetAnswered(time.Unix(1613387999, 0)).SetChoice(choices[0]).SetAnswerer(players[2]),
		tx.Answer.Create().SetID(uuid.MustParse("e26a530e-48ce-4268-8f84-cfe661e2a32a")).SetAnswered(time.Unix(1613388000, 0)).SetChoice(choices[2]).SetAnswerer(players[4]),
	}

	answers := tx.Answer.CreateBulk(answersC...).SaveX(c)

	_, _ = askedQuestions, answers

	if err := tx.Commit(); err != nil {
		t.Fatalf("Error during default data insertion: %v", err)
	}

	return m
}

func TestModel_NextQuestion(t *testing.T) {
	//t.Parallel() // TODO
	m := newTestModelWithData(t)
	c := context.Background()

	if err := m.NextQuestion(uuid.MustParse("b3d2f5b2-d5eb-4461-b352-622431a35b12"), time.Unix(1613388006, 0), c); err != nil {
		t.Fatalf("Unexpected error when switching to next question: %v", err)
	}
}

func TestModel_NextQuestion_noNextQuestion(t *testing.T) {
	//t.Parallel() // TODO
	m := newTestModelWithData(t)
	c := context.Background()

	if err := m.NextQuestion(uuid.MustParse("b3d2f5b2-d5eb-4461-b352-622431a35b12"), time.Unix(1613388006, 0), c); err != nil {
		t.Fatalf("Unexpected error when switching to next question: %v", err)
	}

	if err := m.NextQuestion(uuid.MustParse("b3d2f5b2-d5eb-4461-b352-622431a35b12"), time.Unix(1613388008, 0), c); err == nil {
		t.Fatalf("Switching to next question from the last one did not fail")
	} else if !errors.Is(err, NoNextQuestion) {
		t.Fatalf("Unexpected error type after switching to next question from the last one: %v", err)
	}
}
