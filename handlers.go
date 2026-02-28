package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

type PageData struct {
	Verb        string
	Translation string
	Tenses      []Tense
	Error       string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := PageData{}

	if r.Method == http.MethodPost {
		raw := r.FormValue("verb")

		if strings.TrimSpace(raw) == "" {
			data.Error = "Anna verbi."
			tmpl.Execute(w, data)
			return
		}

		verb, ok := findVerb(raw)
		data.Verb = verb

		if ok {
			data.Tenses = allTenses(verb)
			data.Translation = verbs[verb].Translation
		} else {
			data.Error = "Verbiä ei löydy. Kokeile esimerkiksi: laulaa, mennä, syödä."
		}
	}

	tmpl.Execute(w, data)
}

func verbsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	names := make([]string, 0, len(verbs))
	for k := range verbs {
		names = append(names, k)
	}
	json.NewEncoder(w).Encode(names)
}
