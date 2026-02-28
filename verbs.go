package main

import (
	"encoding/json"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// VerbForms holds all conjugation data for a single verb.
// Every script that reads/writes verbs.json must use this exact struct.
type VerbForms struct {
	Translation         string   `json:"translation"`
	Preesens            []string `json:"preesens"`
	PreesensNeg         []string `json:"preesens_neg"`
	Imperfekti          []string `json:"imperfekti"`
	ImperfektiNeg       []string `json:"imperfekti_neg"`
	Perfekti            []string `json:"perfekti"`
	PerfektiNeg         []string `json:"perfekti_neg"`
	Pluskvamperfekti    []string `json:"pluskvamperfekti"`
	PluskvamperfektiNeg []string `json:"pluskvamperfekti_neg"`
	Konditionaali       []string `json:"konditionaali"`
	KonditionaaliNeg    []string `json:"konditionaali_neg"`
	Imperatiivi         []string `json:"imperatiivi"`
	ImperatiiviNeg      []string `json:"imperatiivi_neg"`
}

type ConjugationRow struct {
	Pronoun string
	Form    string
}

type Tense struct {
	Name string
	Rows []ConjugationRow
}

var verbs map[string]VerbForms

var pronouns = []string{"minä", "sinä", "hän/se", "me", "te", "Te", "he", "passiivi"}

func loadVerbs() {
	data, err := os.ReadFile("verbs.json")
	if err != nil {
		panic("could not read verbs.json: " + err.Error())
	}
	if err := json.Unmarshal(data, &verbs); err != nil {
		panic("could not parse verbs.json: " + err.Error())
	}
}

// safeRows ensures we always have exactly len(pronouns) entries.
func safeRows(rows []string) []string {
	out := make([]string, len(pronouns))
	copy(out, rows)
	return out
}

func allTenses(verb string) []Tense {
	f := verbs[verb]

	tenseNames := []string{
		"Preesens",
		"Preesens, negatiivinen",
		"Imperfekti",
		"Imperfekti, negatiivinen",
		"Perfekti",
		"Perfekti, negatiivinen",
		"Pluskvamperfekti",
		"Pluskvamperfekti, negatiivinen",
	}

	rawForms := [][]string{
		f.Preesens,
		f.PreesensNeg,
		f.Imperfekti,
		f.ImperfektiNeg,
		f.Perfekti,
		f.PerfektiNeg,
		f.Pluskvamperfekti,
		f.PluskvamperfektiNeg,
	}

	tenses := make([]Tense, len(tenseNames))
	for i, name := range tenseNames {
		tenses[i].Name = name
		rows := safeRows(rawForms[i])
		for j, pronoun := range pronouns {
			tenses[i].Rows = append(tenses[i].Rows, ConjugationRow{
				Pronoun: pronoun,
				Form:    rows[j],
			})
		}
	}

	return tenses
}

// normalizeVerbInput cleans user input for lookup.
func normalizeVerbInput(s string) string {
	s = norm.NFC.String(s)
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func findVerb(raw string) (string, bool) {
	norm := normalizeVerbInput(raw)
	if norm == "" {
		return "", false
	}
	_, ok := verbs[norm]
	return norm, ok
}
