package gameCreator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestCreateTemplate(t *testing.T) {
	const expected = "Nadpis otázky,10000,\n,Nadpis možnosti,1\n,Nadpis možnosti,\n,Nadpis možnosti,\n,Nadpis možnosti,\n" +
		"Nadpis otázky,,\n,Nadpis možnosti,1\n,Nadpis možnosti,\n,Nadpis možnosti,\n,Nadpis možnosti,\n" +
		"Nadpis otázky,,\n,Nadpis možnosti,1\n,Nadpis možnosti,\n,Nadpis možnosti,\n,Nadpis možnosti,\n" +
		"Nadpis otázky,,\n,Nadpis možnosti,1\n,Nadpis možnosti,\n,Nadpis možnosti,\n,Nadpis možnosti,\n" +
		"Nadpis otázky,,\n,Nadpis možnosti,1\n,Nadpis možnosti,\n,Nadpis možnosti,\n,Nadpis možnosti,\n"
	var actual bytes.Buffer
	actual.Grow(len(expected))
	if err := CreateTemplate(&actual, 5, 4); err != nil {
		t.Fatalf("Unexpected error returned from CreateTemplate: %v", err)
	}
	if act := actual.String(); act != expected {
		t.Fatalf("Wrong template generated. Expected:\n%s\n\nGot:\n%s\n", expected, act)
	}
}
func TestParse(t *testing.T) {
	const input = "H2O is,3000,\n,Gasoline,\n,Salt,\n,Water,1\n" +
		"π is rational,,\n,Yes,\n,No,1\n" +
		"IPv4 address length is,5000,\n,8b,\n,16b,\n,32b,1\n,64b,\n,128b,\n"
	var expected = Game{
		Questions: []Question{
			{
				Title:  "H2O is",
				Length: 3000,
				Choices: []Choice{
					{Title: "Gasoline"},
					{Title: "Salt"},
					{Title: "Water", Correct: true},
				},
			},
			{
				Title:  "π is rational",
				Length: 3000,
				Choices: []Choice{
					{Title: "Yes"},
					{Title: "No", Correct: true},
				},
			},
			{
				Title:  "IPv4 address length is",
				Length: 5000,
				Choices: []Choice{
					{Title: "8b"},
					{Title: "16b"},
					{Title: "32b", Correct: true},
					{Title: "64b"},
					{Title: "128b"},
				},
			},
		},
	}
	if g, err := Parse(strings.NewReader(input), 10, 10); err == nil && !reflect.DeepEqual(g, expected) {
		t.Fatalf("Parse:\n\tActual: %#v\n\tExpected: %#v", g, expected)
	} else if err != nil {
		t.Fatalf("Unexpected error from Parse: %v", err)
	}
}
