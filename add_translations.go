//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// VerbForms must match the struct in verbs.go exactly.
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

var httpClient = &http.Client{}

func get(apiURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FinnishConjugatorBot/1.0 (your@email.com)")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func fetchTranslation(verb string) (string, error) {
	body, err := get(fmt.Sprintf(
		"https://en.wiktionary.org/w/api.php?action=parse&page=%s&prop=wikitext&format=json",
		url.QueryEscape(verb),
	))
	if err != nil {
		return "", err
	}

	var wiki struct {
		Parse struct {
			Wikitext struct {
				Content string `json:"*"`
			} `json:"wikitext"`
		} `json:"parse"`
	}
	if err := json.Unmarshal(body, &wiki); err != nil {
		return "", err
	}

	wikitext := wiki.Parse.Wikitext.Content
	if wikitext == "" {
		return "", fmt.Errorf("empty wikitext")
	}

	re := regexp.MustCompile(`(?m)^# (?:\{\{[^}]+\}\} )?(.+)`)
	match := re.FindStringSubmatch(wikitext)
	if len(match) < 2 {
		return "", fmt.Errorf("no definition found")
	}

	def := match[1]
	def = regexp.MustCompile(`\[\[(?:[^\]|]+\|)?([^\]]+)\]\]`).ReplaceAllString(def, "$1")
	def = regexp.MustCompile(`\{\{[^}]+\}\}`).ReplaceAllString(def, "")
	return strings.TrimSpace(def), nil
}

func main() {
	data, err := os.ReadFile("verbs.json")
	if err != nil {
		panic(err)
	}

	var verbs map[string]VerbForms
	if err := json.Unmarshal(data, &verbs); err != nil {
		panic(err)
	}

	total := len(verbs)
	i, updated, failed := 0, 0, 0

	for verb, forms := range verbs {
		i++
		if forms.Translation != "" {
			fmt.Printf("[%d/%d] skipping %s (already has translation)\n", i, total, verb)
			continue
		}

		fmt.Printf("[%d/%d] fetching translation for %s...\n", i, total, verb)

		translation, err := fetchTranslation(verb)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			failed++
			continue
		}

		forms.Translation = translation
		verbs[verb] = forms
		updated++
		fmt.Printf("  → %s\n", translation)

		if updated%10 == 0 {
			out, _ := json.MarshalIndent(verbs, "", "    ")
			os.WriteFile("verbs.json", out, 0644)
			fmt.Printf("  saved progress (%d updated)\n", updated)
		}

		time.Sleep(500 * time.Millisecond)
	}

	out, _ := json.MarshalIndent(verbs, "", "    ")
	os.WriteFile("verbs.json", out, 0644)
	fmt.Printf("\nDone! Updated: %d  Failed: %d\n", updated, failed)
}
