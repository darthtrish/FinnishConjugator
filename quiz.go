package main

import (
	"html/template"
	"math/rand"
	"net/http"
	"strings"
)

var quizTmpl = template.Must(template.ParseFiles("templates/quiz.html"))

type QuizData struct {
	Verb    string
	Tense   string
	Pronoun string
	Answer  string
	Input   string
	Result  string // "correct", "wrong", or ""
}

var quizTenseNames = []string{
	"Preesens",
	"Preesens, negatiivinen",
	"Imperfekti",
	"Imperfekti, negatiivinen",
	"Perfekti",
	"Perfekti, negatiivinen",
	"Pluskvamperfekti",
	"Pluskvamperfekti, negatiivinen",
}

var quizPronouns = []string{
	"minä", "sinä", "hän/se", "me", "te", "he", "passiivi",
}

func randomQuiz() QuizData {
	keys := make([]string, 0, len(verbs))
	for k := range verbs {
		keys = append(keys, k)
	}
	verb := keys[rand.Intn(len(keys))]
	tense := quizTenseNames[rand.Intn(len(quizTenseNames))]
	pronoun := quizPronouns[rand.Intn(len(quizPronouns))]

	answer := ""
	for _, t := range allTenses(verb) {
		if t.Name == tense {
			for _, row := range t.Rows {
				if row.Pronoun == pronoun {
					answer = row.Form
					break
				}
			}
			break
		}
	}

	return QuizData{
		Verb:    verb,
		Tense:   tense,
		Pronoun: pronoun,
		Answer:  answer,
	}
}

func quizHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodGet {
		quizTmpl.Execute(w, randomQuiz())
		return
	}

	// POST — check the answer
	answer := r.FormValue("answer")
	userInput := strings.TrimSpace(strings.ToLower(r.FormValue("input")))
	correct := strings.TrimSpace(strings.ToLower(answer))

	result := "wrong"
	if userInput == correct {
		result = "correct"
	}

	quizTmpl.Execute(w, QuizData{
		Verb:    r.FormValue("verb"),
		Tense:   r.FormValue("tense"),
		Pronoun: r.FormValue("pronoun"),
		Answer:  answer,
		Input:   r.FormValue("input"),
		Result:  result,
	})
}
