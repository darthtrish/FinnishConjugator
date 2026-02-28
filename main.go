package main

import (
	"net/http"
)

func main() {
	loadVerbs()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/verbs", verbsHandler)
	http.HandleFunc("/quiz", quizHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
