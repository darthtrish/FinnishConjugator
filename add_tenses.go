//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

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

type WikiResponse struct {
	Parse struct {
		Wikitext struct {
			Content string `json:"*"`
		} `json:"wikitext"`
	} `json:"parse"`
}

type ExpandResponse struct {
	Expandtemplates struct {
		Wikitext string `json:"wikitext"`
	} `json:"expandtemplates"`
}

var client = &http.Client{}

func get(apiUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FinnishConjugatorBot/1.0 (your@email.com)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func fetchExpanded(verb string) (string, error) {
	body, err := get(fmt.Sprintf(
		"https://en.wiktionary.org/w/api.php?action=parse&page=%s&prop=wikitext&format=json",
		url.QueryEscape(verb),
	))
	if err != nil {
		return "", err
	}

	var wiki WikiResponse
	if err := json.Unmarshal(body, &wiki); err != nil {
		return "", err
	}

	wikitext := wiki.Parse.Wikitext.Content
	re := regexp.MustCompile(`\{\{fi-conj[^}]+\}\}`)
	match := re.FindString(wikitext)
	if match == "" {
		return "", fmt.Errorf("no fi-conj template found")
	}

	body2, err := get(
		"https://en.wiktionary.org/w/api.php?action=expandtemplates&prop=wikitext&format=json&text=" +
			url.QueryEscape(match),
	)
	if err != nil {
		return "", err
	}

	var expand ExpandResponse
	if err := json.Unmarshal(body2, &expand); err != nil {
		return "", err
	}

	return expand.Expandtemplates.Wikitext, nil
}

func extractForms(expanded string) []string {
	cellRe := regexp.MustCompile(`data-accel-col="\d+" \| ([^\n]+)`)
	linkRe := regexp.MustCompile(`\[\[:([^#\]]+)#Finnish\|([^\]]+)\]\]`)
	tagRe := regexp.MustCompile(`<[^>]+>`)

	var forms []string
	for _, m := range cellRe.FindAllStringSubmatch(expanded, -1) {
		cell := m[1]
		cell = linkRe.ReplaceAllString(cell, "$2")
		cell = tagRe.ReplaceAllString(cell, "")
		cell = strings.TrimSpace(cell)
		if cell != "" {
			forms = append(forms, cell)
		}
	}
	return forms
}

func extractNewTenses(expanded string) (konditionaali, konditionaaliNeg, imperatiivi, imperatiiviNeg []string, ok bool) {
	forms := extractForms(expanded)

	fmt.Printf("  total forms: %d\n", len(forms))
	// Print forms around where imperatiivi should be
	for i := 80; i < len(forms) && i < 120; i++ {
		fmt.Printf("  [%d] %s\n", i, forms[i])
	}

	if len(forms) < 84 {
		return nil, nil, nil, nil, false
	}

	get := func(block, row, col int) string {
		idx := block*28 + row*4 + col
		if idx < len(forms) {
			return forms[idx]
		}
		return ""
	}

	build := func(block, col int) []string {
		return []string{
			get(block, 0, col), get(block, 1, col), get(block, 2, col),
			get(block, 3, col), get(block, 4, col), get(block, 4, col),
			get(block, 5, col), get(block, 6, col),
		}
	}

	return build(2, 0), build(2, 1), build(3, 0), build(3, 1), true
}

func main() {
	expanded, err := fetchExpanded("laulaa")
	if err != nil {
		panic(err)
	}
	extractNewTenses(expanded)
}
