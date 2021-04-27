package gameCreator

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"time"
)

func CreateTemplate(w io.Writer, questions uint64, choicesPerQuestion uint64) (retE error) {
	csvW := csv.NewWriter(w)
	defer func() {
		csvW.Flush()
		if err := csvW.Error(); err != nil && retE != nil {
			retE = err
		}
	}()

	for i := uint64(0); i < questions; i++ {
		var length string
		if i == 0 {
			length = "10000"
		}
		if err := csvW.Write([]string{"Nadpis otázky", length, ""}); err != nil {
			return err
		}
		for j := uint64(0); j < choicesPerQuestion; j++ {
			var correct string
			if j == 0 {
				correct = "1"
			}
			if err := csvW.Write([]string{"", "Nadpis možnosti", correct}); err != nil {
				return err
			}
		}
	}
	return nil
}

type Game struct {
	Questions []Question
}

type Question struct {
	Title   string
	Choices []Choice
	Length  uint64
}

type Choice struct {
	Title   string
	Correct bool
}

var ErrTooManyQuestions = errors.New("there were questions above the limit")
var ErrTooManyChoices = errors.New("there were choices above the limit")
var ErrInvalidSyntax = errors.New("")

func Parse(r io.Reader, maxQuestions uint64, maxChoicesPerQuestion uint64) (Game, error) {
	var g Game
	var csvR = csv.NewReader(r)
	csvR.FieldsPerRecord = 3
	csvR.TrimLeadingSpace = true
	var questions, choices uint64
	for {
		if row, err := csvR.Read(); err == nil {
			if row[0] == "" && row[1] == "" && row[2] == "" {
				continue
			} else if row[0] == "" {
				choices++
				if questions == 0 {
					return g, ErrInvalidSyntax
				}
				if choices > maxChoicesPerQuestion {
					return g, ErrTooManyChoices
				}
				var correct bool
				if row[2] == "1" {
					correct = true
				}
				g.Questions[len(g.Questions)-1].Choices = append(g.Questions[len(g.Questions)-1].Choices, Choice{
					Title:   row[1],
					Correct: correct,
				})
			} else {
				questions++
				choices = 0
				if questions > maxQuestions {
					return g, ErrTooManyQuestions
				}
				var length uint64
				if row[1] != "" {
					if l, err := strconv.ParseUint(row[1], 10, 64); err == nil {
						length = l
					} else {
						return g, ErrInvalidSyntax
					}
				} else {
					if questions > 1 {
						length = g.Questions[len(g.Questions)-1].Length
					} else {
						length = uint64((10 * time.Second).Milliseconds())
					}
				}
				g.Questions = append(g.Questions, Question{
					Title:  row[0],
					Length: length,
				})
			}
		} else if err == io.EOF {
			break
		} else if err == csv.ErrFieldCount {
			return g, ErrInvalidSyntax
		} else {
			return g, err
		}
	}
	return g, nil
}
