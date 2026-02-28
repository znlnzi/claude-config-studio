package evolution

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

// Analyzer is the local analysis engine (no LLM dependency)
type Analyzer struct {
	claudeDir string
	scope     string
}

// NewAnalyzer creates an Analyzer instance
func NewAnalyzer(claudeDir, scope string) *Analyzer {
	return &Analyzer{claudeDir: claudeDir, scope: scope}
}

// ruleFile represents a single rule file
type ruleFile struct {
	Name     string
	Path     string
	Content  string
	Size     int64
	Keywords map[string]int
}

// Analyze performs a full local analysis
func (a *Analyzer) Analyze() ([]Suggestion, AnalysisRecord, error) {
	start := time.Now()

	rules, err := a.loadRules()
	if err != nil {
		return nil, AnalysisRecord{}, err
	}

	var suggestions []Suggestion

	// 1. Duplicate detection
	suggestions = append(suggestions, a.findDuplicates(rules)...)

	// 2. Coverage check
	suggestions = append(suggestions, a.checkCoverage(rules)...)

	// 3. Health check
	suggestions = append(suggestions, a.checkHealth(rules)...)

	record := AnalysisRecord{
		ID:               NewSuggestionID(),
		Scope:            a.scope,
		Timestamp:        time.Now().Format(time.RFC3339),
		RulesScanned:     len(rules),
		SuggestionsFound: len(suggestions),
		DurationMs:       time.Since(start).Milliseconds(),
	}

	return suggestions, record, nil
}

// loadRules loads all rule files
func (a *Analyzer) loadRules() ([]ruleFile, error) {
	rulesDir := filepath.Join(a.claudeDir, "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, nil
	}

	var rules []ruleFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(rulesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		info, _ := entry.Info()
		name := strings.TrimSuffix(entry.Name(), ".md")
		content := string(data)

		rules = append(rules, ruleFile{
			Name:     name,
			Path:     path,
			Content:  content,
			Size:     info.Size(),
			Keywords: extractKeywords(content),
		})
	}

	return rules, nil
}

// findDuplicates detects duplicate rules based on keyword overlap
func (a *Analyzer) findDuplicates(rules []ruleFile) []Suggestion {
	var suggestions []Suggestion

	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			overlap := keywordOverlap(rules[i].Keywords, rules[j].Keywords)
			if overlap > 0.6 {
				suggestions = append(suggestions, Suggestion{
					ID:    NewSuggestionID(),
					Type:  "duplicate_rules",
					Title: "Highly similar rule files detected",
					Description: "Files " + rules[i].Name + ".md and " + rules[j].Name + ".md " +
						"have a keyword overlap of " + formatPercent(overlap) + ", possibly containing duplicate content.",
					Files:      []string{rules[i].Name + ".md", rules[j].Name + ".md"},
					Suggestion: "Review both files and consider merging duplicate content or clarifying their respective responsibilities.",
					Confidence: overlap,
					Source:     "local_analysis",
					CreatedAt:  time.Now().Format(time.RFC3339),
					Status:     "pending",
				})
			}
			// Brief delay to avoid ID collisions
			time.Sleep(time.Millisecond)
		}
	}

	return suggestions
}

// checkCoverage checks whether key rule categories are missing
func (a *Analyzer) checkCoverage(rules []ruleFile) []Suggestion {
	// Expected key rule categories
	expectedCategories := map[string][]string{
		"Security":       {"security", "safe", "xss", "sql injection", "owasp"},
		"Testing":        {"test", "testing", "tdd", "coverage", "rate"},
		"Error Handling": {"error", "fault", "exception", "abnormal"},
		"Code Style":     {"style", "coding style", "lint", "format", "naming"},
		"Git":            {"git", "commit", "branch", "pr", "merge"},
	}

	// Collect keywords from all rules
	allKeywords := make(map[string]int)
	for _, rule := range rules {
		for kw, count := range rule.Keywords {
			allKeywords[kw] += count
		}
	}

	var suggestions []Suggestion
	for category, keywords := range expectedCategories {
		found := false
		for _, kw := range keywords {
			if allKeywords[kw] > 0 {
				found = true
				break
			}
		}
		if !found {
			suggestions = append(suggestions, Suggestion{
				ID:          NewSuggestionID(),
				Type:        "missing_coverage",
				Title:       "Missing " + category + " rules",
				Description: "No " + category + " related content found in current rules. Consider adding corresponding rules to improve configuration completeness.",
				Suggestion:  "Consider adding a rule file related to " + category + ".",
				Confidence:  0.7,
				Source:      "local_analysis",
				CreatedAt:   time.Now().Format(time.RFC3339),
				Status:      "pending",
			})
			time.Sleep(time.Millisecond)
		}
	}

	return suggestions
}

// checkHealth checks the health of rule files
func (a *Analyzer) checkHealth(rules []ruleFile) []Suggestion {
	var suggestions []Suggestion

	for _, rule := range rules {
		// File too large (> 5KB)
		if rule.Size > 5120 {
			suggestions = append(suggestions, Suggestion{
				ID:          NewSuggestionID(),
				Type:        "health_issue",
				Title:       rule.Name + ".md file is too large",
				Description: "File size is " + formatSize(rule.Size) + ", exceeding the recommended 5KB limit. Oversized rule files may waste context window space.",
				Files:       []string{rule.Name + ".md"},
				Suggestion:  "Consider splitting the file into smaller files, each focusing on a single topic.",
				Confidence:  0.75,
				Source:      "local_analysis",
				CreatedAt:   time.Now().Format(time.RFC3339),
				Status:      "pending",
			})
			time.Sleep(time.Millisecond)
		}

		// File too small (< 100B, possibly an empty template)
		if rule.Size < 100 && rule.Size > 0 {
			suggestions = append(suggestions, Suggestion{
				ID:          NewSuggestionID(),
				Type:        "health_issue",
				Title:       rule.Name + ".md file has too little content",
				Description: "File is only " + formatSize(rule.Size) + ", possibly an incomplete template or placeholder file.",
				Files:       []string{rule.Name + ".md"},
				Suggestion:  "Consider adding content or removing the empty file.",
				Confidence:  0.6,
				Source:      "local_analysis",
				CreatedAt:   time.Now().Format(time.RFC3339),
				Status:      "pending",
			})
			time.Sleep(time.Millisecond)
		}

		// Missing heading (Markdown files should start with #)
		trimmed := strings.TrimSpace(rule.Content)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "<!--") {
			suggestions = append(suggestions, Suggestion{
				ID:          NewSuggestionID(),
				Type:        "health_issue",
				Title:       rule.Name + ".md is missing a heading",
				Description: "Markdown rule files should start with a heading (#) for easy identification and indexing.",
				Files:       []string{rule.Name + ".md"},
				Suggestion:  "Consider adding a descriptive heading at the beginning of the file.",
				Confidence:  0.65,
				Source:      "local_analysis",
				CreatedAt:   time.Now().Format(time.RFC3339),
				Status:      "pending",
			})
			time.Sleep(time.Millisecond)
		}
	}

	return suggestions
}

// ─── Helper Functions ─────────────────────────────

// extractKeywords extracts keywords and their frequencies from text
func extractKeywords(text string) map[string]int {
	keywords := make(map[string]int)
	text = strings.ToLower(text)

	// Simple tokenization: split by non-alphanumeric characters
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	// Filter stop words and words that are too short
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "and": true, "or": true, "not": true, "with": true,
		"this": true, "that": true, "it": true, "be": true, "as": true,
		"from": true, "by": true, "was": true, "were": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"will": true, "can": true, "use": true, "using": true,
	}

	for _, w := range words {
		if len(w) < 2 || stopWords[w] {
			continue
		}
		keywords[w]++
	}

	return keywords
}

// keywordOverlap computes the Jaccard similarity of two keyword sets
func keywordOverlap(a, b map[string]int) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	// Compare top N keywords (to avoid dilution from large vocabulary in long files)
	topA := topKeywords(a, 30)
	topB := topKeywords(b, 30)

	setA := make(map[string]bool)
	for _, kw := range topA {
		setA[kw] = true
	}

	intersection := 0
	for _, kw := range topB {
		if setA[kw] {
			intersection++
		}
	}

	union := len(topA) + len(topB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// topKeywords returns the top N keywords by frequency
func topKeywords(keywords map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range keywords {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	var result []string
	for i, item := range sorted {
		if i >= n {
			break
		}
		result = append(result, item.Key)
	}
	return result
}

func formatPercent(f float64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(
			strings.Replace(
				strings.Replace(
					formatFloat(f*100), ".", ".", 1,
				), ",", "", -1,
			), " ", "", -1,
		), "0"), ".") + "%"
}

func formatFloat(f float64) string {
	s := strings.TrimRight(strings.TrimRight(
		strings.Replace(
			strings.Replace(
				strings.Replace(
					func() string {
						return strings.Replace(
							strings.TrimRight(strings.TrimRight(
								func() string {
									r := make([]byte, 0, 10)
									i := int(f)
									frac := f - float64(i)
									if frac < 0 {
										frac = -frac
									}
									r = append(r, []byte(intToStr(i))...)
									r = append(r, '.')
									frac10 := int(frac*10 + 0.5)
									r = append(r, byte('0'+frac10%10))
									return string(r)
								}(), "0"), "."),
							".", ".", 1)
					}(), ".", ".", 1,
				), ",", "", -1,
			), " ", "", -1,
		), "0"), ".")
	return s
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return intToStr(int(bytes)) + "B"
	}
	return intToStr(int(bytes/1024)) + "KB"
}
