package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

type VerbForms struct {
	Preesens            []string `json:"preesens"`
	PreesensNeg         []string `json:"preesens_neg"`
	Imperfekti          []string `json:"imperfekti"`
	ImperfektiNeg       []string `json:"imperfekti_neg"`
	Perfekti            []string `json:"perfekti"`
	PerfektiNeg         []string `json:"perfekti_neg"`
	Pluskvamperfekti    []string `json:"pluskvamperfekti"`
	PluskvamperfektiNeg []string `json:"pluskvamperfekti_neg"`
}

type ConjugationRow struct {
	Pronoun string
	Form    string
}

type Tense struct {
	Name string
	Rows []ConjugationRow
}

type PageData struct {
	Verb   string
	Tenses []Tense
	Error  string
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

// Always return exactly len(pronouns) items
func safeRows(rows []string) []string {
	out := make([]string, len(pronouns))
	copy(out, rows)
	return out
}

func allTenses(verb string) []Tense {
	f := verbs[verb]

	tenses := []Tense{
		{Name: "Preesens"},
		{Name: "Preesens, negatiivinen"},
		{Name: "Imperfekti"},
		{Name: "Imperfekti, negatiivinen"},
		{Name: "Perfekti"},
		{Name: "Perfekti, negatiivinen"},
		{Name: "Pluskvamperfekti"},
		{Name: "Pluskvamperfekti, negatiivinen"},
	}

	forms := [][]string{
		f.Preesens,
		f.PreesensNeg,
		f.Imperfekti,
		f.ImperfektiNeg,
		f.Perfekti,
		f.PerfektiNeg,
		f.Pluskvamperfekti,
		f.PluskvamperfektiNeg,
	}

	for i, rows := range forms {
		rows = safeRows(rows)
		for j, pronoun := range pronouns {
			tenses[i].Rows = append(
				tenses[i].Rows,
				ConjugationRow{Pronoun: pronoun, Form: rows[j]},
			)
		}
	}

	return tenses
}

// --- INPUT NORMALIZATION ---

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
	normVerb := normalizeVerbInput(raw)
	if normVerb == "" {
		return "", false
	}
	_, ok := verbs[normVerb]
	return normVerb, ok
}

// --- HTTP ---

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := PageData{}

	if r.Method == http.MethodPost {
		raw := r.FormValue("verb")

		if strings.TrimSpace(raw) == "" {
			data.Error = "Anna verbi."
			_ = tmpl.Execute(w, data)
			return
		}

		normVerb, ok := findVerb(raw)
		data.Verb = normVerb

		if ok {
			data.Tenses = allTenses(normVerb)
		} else {
			data.Error = "Verbiä ei löydy. Kokeile esimerkiksi: laulaa, mennä, syödä."
		}
	}

	_ = tmpl.Execute(w, data)
}

func main() {
	loadVerbs()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
