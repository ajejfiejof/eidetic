package index

import (
	"strings"
	"unicode"
)

// Tokenize extracts normalized tokens from text with code and shell awareness.
func Tokenize(text string) []string {
	if text == "" {
		return nil
	}

	var tokens []string
	seen := make(map[string]struct{})

	addToken := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if len(t) < 2 {
			return
		}
		if _, exists := seen[t]; !exists {
			seen[t] = struct{}{}
			tokens = append(tokens, t)
		}
	}

	// Split by whitespace first
	fields := strings.Fields(text)
	for _, field := range fields {
		// Clean outer punctuation
		clean := strings.Trim(field, " \t\n\r`'\"()[]{}<>,;:")
		if clean == "" {
			continue
		}

		addToken(clean)

		// Path splitting (/path/to/file)
		if strings.Contains(clean, "/") || strings.Contains(clean, "\\") {
			parts := strings.FieldsFunc(clean, func(r rune) bool {
				return r == '/' || r == '\\' || r == '.'
			})
			for _, p := range parts {
				addToken(p)
			}
		}

		// Delimiter splitting: snake_case, kebab-case, flags
		if strings.ContainsAny(clean, "_-:=.") {
			subParts := strings.FieldsFunc(clean, func(r rune) bool {
				return r == '_' || r == '-' || r == ':' || r == '=' || r == '.'
			})
			for _, sp := range subParts {
				addToken(sp)
			}
		}

		// CamelCase / PascalCase splitting
		var currentWord strings.Builder
		for i, r := range clean {
			if unicode.IsUpper(r) {
				if currentWord.Len() > 0 && (i+1 < len(clean) && unicode.IsLower(rune(clean[i+1]))) {
					addToken(currentWord.String())
					currentWord.Reset()
				}
			}
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				currentWord.WriteRune(r)
			} else {
				if currentWord.Len() > 0 {
					addToken(currentWord.String())
					currentWord.Reset()
				}
			}
		}
		if currentWord.Len() > 0 {
			addToken(currentWord.String())
		}
	}

	return tokens
}

// GenerateTrigrams produces 3-character n-grams for fuzzy matching.
func GenerateTrigrams(text string) []string {
	normalized := strings.ToLower(strings.TrimSpace(text))
	runes := []rune(normalized)
	if len(runes) < 3 {
		return []string{normalized}
	}

	var trigrams []string
	seen := make(map[string]struct{})
	for i := 0; i <= len(runes)-3; i++ {
		tri := string(runes[i : i+3])
		if _, exists := seen[tri]; !exists {
			seen[tri] = struct{}{}
			trigrams = append(trigrams, tri)
		}
	}
	return trigrams
}
