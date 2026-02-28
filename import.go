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

type wikiResponse struct {
	Parse struct {
		Wikitext struct {
			Content string `json:"*"`
		} `json:"wikitext"`
	} `json:"parse"`
}

type expandResponse struct {
	Expandtemplates struct {
		Wikitext string `json:"wikitext"`
	} `json:"expandtemplates"`
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

func fetchExpanded(verb string) (expanded, wikitext string, err error) {
	body, err := get(fmt.Sprintf(
		"https://en.wiktionary.org/w/api.php?action=parse&page=%s&prop=wikitext&format=json",
		url.QueryEscape(verb),
	))
	if err != nil {
		return "", "", err
	}

	var wiki wikiResponse
	if err := json.Unmarshal(body, &wiki); err != nil {
		return "", "", err
	}

	wikitext = wiki.Parse.Wikitext.Content
	re := regexp.MustCompile(`\{\{fi-conj[^}]+\}\}`)
	match := re.FindString(wikitext)
	if match == "" {
		return "", "", fmt.Errorf("no fi-conj template found")
	}

	body2, err := get(
		"https://en.wiktionary.org/w/api.php?action=expandtemplates&prop=wikitext&format=json&text=" +
			url.QueryEscape(match),
	)
	if err != nil {
		return "", "", err
	}

	var expand expandResponse
	if err := json.Unmarshal(body2, &expand); err != nil {
		return "", "", err
	}

	return expand.Expandtemplates.Wikitext, wikitext, nil
}

func extractTranslation(wikitext string) string {
	re := regexp.MustCompile(`(?m)^# (?:\{\{[^}]+\}\} )?(.+)`)
	match := re.FindStringSubmatch(wikitext)
	if len(match) < 2 {
		return ""
	}
	def := match[1]
	def = regexp.MustCompile(`\[\[(?:[^\]|]+\|)?([^\]]+)\]\]`).ReplaceAllString(def, "$1")
	def = regexp.MustCompile(`\{\{[^}]+\}\}`).ReplaceAllString(def, "")
	return strings.TrimSpace(def)
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

func parseVerb(expanded string) (VerbForms, bool) {
	forms := extractForms(expanded)

	if len(forms) < 56 {
		fmt.Printf("  not enough forms: got %d, need 56\n", len(forms))
		return VerbForms{}, false
	}

	at := func(block, row, col int) string {
		idx := block*28 + row*4 + col
		if idx < len(forms) {
			return forms[idx]
		}
		return ""
	}

	build := func(block, col int) []string {
		return []string{
			at(block, 0, col), at(block, 1, col), at(block, 2, col),
			at(block, 3, col), at(block, 4, col), at(block, 4, col),
			at(block, 5, col), at(block, 6, col),
		}
	}

	return VerbForms{
		Preesens:            build(0, 0),
		PreesensNeg:         build(0, 1),
		Perfekti:            build(0, 2),
		PerfektiNeg:         build(0, 3),
		Imperfekti:          build(1, 0),
		ImperfektiNeg:       build(1, 1),
		Pluskvamperfekti:    build(1, 2),
		PluskvamperfektiNeg: build(1, 3),
	}, true
}

func loadExisting(path string) map[string]VerbForms {
	existing := map[string]VerbForms{}
	data, err := os.ReadFile(path)
	if err != nil {
		return existing
	}
	json.Unmarshal(data, &existing)
	return existing
}

func saveVerbs(path string, data map[string]VerbForms) error {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func readVerbList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var verbs []string
	for _, line := range strings.Split(string(data), "\n") {
		v := strings.TrimSpace(line)
		if v != "" {
			verbs = append(verbs, v)
		}
	}
	return verbs, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run import.go verbs_to_import.txt")
		fmt.Println("       go run import.go laulaa mennä tulla")
		os.Exit(1)
	}

	var verbList []string
	if _, err := os.Stat(os.Args[1]); err == nil {
		verbList, _ = readVerbList(os.Args[1])
	} else {
		verbList = os.Args[1:]
	}

	existing := loadExisting("verbs.json")
	added, skipped, failed := 0, 0, 0

	for i, verb := range verbList {
		if _, ok := existing[verb]; ok {
			fmt.Printf("[%d/%d] skipping %s (already exists)\n", i+1, len(verbList), verb)
			skipped++
			continue
		}

		fmt.Printf("[%d/%d] fetching %s...\n", i+1, len(verbList), verb)

		expanded, wikitext, err := fetchExpanded(verb)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			failed++
			continue
		}

		forms, ok := parseVerb(expanded)
		if !ok {
			fmt.Printf("  ERROR: could not parse forms\n")
			failed++
			continue
		}

		forms.Translation = extractTranslation(wikitext)
		existing[verb] = forms
		added++

		if added%10 == 0 {
			saveVerbs("verbs.json", existing)
			fmt.Printf("  saved progress (%d added)\n", added)
		}

		time.Sleep(500 * time.Millisecond)
	}

	saveVerbs("verbs.json", existing)
	fmt.Printf("\nDone! Added: %d  Skipped: %d  Failed: %d\n", added, skipped, failed)
}
