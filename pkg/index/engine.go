package index

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// SearchFilter specifies constraints on query results.
type SearchFilter struct {
	Source document.SourceType
	Since  time.Duration
	Limit  int
}

// SearchResult contains a matched document with its relevance score and highlight.
type SearchResult struct {
	Doc       document.Document `json:"document"`
	Score     float64           `json:"score"`
	Highlight string            `json:"highlight,omitempty"`
}

// Engine implements an in-memory inverted index with BM25 ranking and trigram fuzzy matching.
type Engine struct {
	mu           sync.RWMutex
	docs         map[string]document.Document
	docLengths   map[string]int
	totalTokens  int
	postings     map[string]map[string]int      // term -> docID -> termFrequency
	trigrams     map[string]map[string]struct{} // trigram -> docID set
	totalDocs    int
	avgDocLength float64
	storage      *Storage
}

// NewEngine initializes the search engine connected to durable storage.
func NewEngine(storage *Storage) (*Engine, error) {
	e := &Engine{
		docs:        make(map[string]document.Document),
		docLengths:  make(map[string]int),
		postings:    make(map[string]map[string]int),
		trigrams:    make(map[string]map[string]struct{}),
		storage:     storage,
	}

	if storage != nil {
		allDocs, err := storage.LoadAll()
		if err != nil {
			return nil, err
		}
		for _, doc := range allDocs {
			e.indexDocument(doc, false)
		}
		e.updateStats()
	}

	return e, nil
}

func (e *Engine) updateStats() {
	e.totalDocs = len(e.docs)
	if e.totalDocs > 0 {
		e.avgDocLength = float64(e.totalTokens) / float64(e.totalDocs)
	} else {
		e.avgDocLength = 0
	}
}

// indexDocument internal indexing logic (caller must hold lock if needed).
func (e *Engine) indexDocument(doc document.Document, persist bool) error {
	if _, exists := e.docs[doc.ID]; exists {
		return nil
	}

	tokens := Tokenize(doc.Content + " " + doc.Title)
	docLen := len(tokens)
	e.docs[doc.ID] = doc
	e.docLengths[doc.ID] = docLen
	e.totalTokens += docLen

	// Calculate term frequencies
	tfMap := make(map[string]int)
	for _, t := range tokens {
		tfMap[t]++
	}

	for term, tf := range tfMap {
		if e.postings[term] == nil {
			e.postings[term] = make(map[string]int)
		}
		e.postings[term][doc.ID] = tf
	}

	// Index trigrams for fuzzy typo recovery
	docTrigrams := GenerateTrigrams(doc.Title + " " + doc.Content)
	for _, tri := range docTrigrams {
		if e.trigrams[tri] == nil {
			e.trigrams[tri] = make(map[string]struct{})
		}
		e.trigrams[tri][doc.ID] = struct{}{}
	}

	if persist && e.storage != nil {
		return e.storage.Append(doc)
	}
	return nil
}

// Index adds a document to memory and storage.
func (e *Engine) Index(doc document.Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.indexDocument(doc, true)
	e.updateStats()
	return err
}

// IndexBatch adds multiple documents efficiently.
func (e *Engine) IndexBatch(docs []document.Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var newDocs []document.Document
	for _, doc := range docs {
		if _, exists := e.docs[doc.ID]; !exists {
			if err := e.indexDocument(doc, false); err == nil {
				newDocs = append(newDocs, doc)
			}
		}
	}

	e.updateStats()

	if e.storage != nil && len(newDocs) > 0 {
		return e.storage.AppendBatch(newDocs)
	}
	return nil
}

// Levenshtein computes minimum edit distance between two strings.
func Levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n1, n2 := len(r1), len(r2)
	if n1 == 0 {
		return n2
	}
	if n2 == 0 {
		return n1
	}

	dp := make([]int, n2+1)
	for j := 0; j <= n2; j++ {
		dp[j] = j
	}

	for i := 1; i <= n1; i++ {
		prev := dp[0]
		dp[0] = i
		for j := 1; j <= n2; j++ {
			temp := dp[j]
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			dp[j] = min(dp[j]+1, min(dp[j-1]+1, prev+cost))
			prev = temp
		}
	}
	return dp[n2]
}

// Search executes BM25 ranking across query tokens with fuzzy recovery.
func (e *Engine) Search(query string, filter SearchFilter) []SearchResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	query = strings.TrimSpace(query)
	if query == "" {
		return e.recentDocuments(filter)
	}

	queryTokens := Tokenize(query)
	if len(queryTokens) == 0 {
		return e.recentDocuments(filter)
	}

	const k1 = 1.2
	const b = 0.75
	scores := make(map[string]float64)

	var cutoffTime time.Time
	if filter.Since > 0 {
		cutoffTime = time.Now().Add(-filter.Since)
	}

	now := time.Now()
	queryLower := strings.ToLower(query)

	for _, token := range queryTokens {
		postings := e.postings[token]

		// Fallback to fuzzy dictionary matching if token has zero exact matches
		if len(postings) == 0 && len(token) >= 3 {
			maxDist := 1
			if len(token) >= 5 {
				maxDist = 2
			}
			for term, termPostings := range e.postings {
				if math.Abs(float64(len(term)-len(token))) <= float64(maxDist) {
					dist := Levenshtein(token, term)
					if dist <= maxDist {
						weight := 1.0 - (float64(dist) / float64(maxDist+1))
						for docID, tf := range termPostings {
							scores[docID] += float64(tf) * weight * 1.5
						}
					}
				}
			}
			continue
		}

		df := float64(len(postings))
		idf := math.Log(1.0 + (float64(e.totalDocs)-df+0.5)/(df+0.5))
		if idf < 0 {
			idf = 0.01
		}

		for docID, tf := range postings {
			docLen := float64(e.docLengths[docID])
			numerator := float64(tf) * (k1 + 1.0)
			denominator := float64(tf) + k1*(1.0-b+b*(docLen/math.Max(1.0, e.avgDocLength)))
			bm25 := idf * (numerator / denominator)
			scores[docID] += bm25
		}
	}

	// Apply source filtering, recency boosts, and substring bonuses
	var results []SearchResult
	for docID, score := range scores {
		doc := e.docs[docID]

		// Source filter
		if filter.Source != "" && doc.Source != filter.Source {
			continue
		}

		// Time cutoff
		if !cutoffTime.IsZero() && doc.Timestamp.Before(cutoffTime) {
			continue
		}

		contentLower := strings.ToLower(doc.Content)
		titleLower := strings.ToLower(doc.Title)

		// Exact match boosts
		if strings.Contains(contentLower, queryLower) {
			score += 15.0
		}
		if strings.Contains(titleLower, queryLower) {
			score += 10.0
		}

		// Recency boost (fresher items get up to +2.0 boost)
		daysOld := now.Sub(doc.Timestamp).Hours() / 24.0
		if daysOld < 0 {
			daysOld = 0
		}
		recencyBoost := 2.0 / (1.0 + math.Log1p(daysOld))
		score += recencyBoost

		results = append(results, SearchResult{
			Doc:   doc,
			Score: score,
		})
	}

	// Sort results descending by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (e *Engine) recentDocuments(filter SearchFilter) []SearchResult {
	var list []document.Document
	var cutoffTime time.Time
	if filter.Since > 0 {
		cutoffTime = time.Now().Add(-filter.Since)
	}

	for _, doc := range e.docs {
		if filter.Source != "" && doc.Source != filter.Source {
			continue
		}
		if !cutoffTime.IsZero() && doc.Timestamp.Before(cutoffTime) {
			continue
		}
		list = append(list, doc)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.After(list[j].Timestamp)
	})

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if len(list) > limit {
		list = list[:limit]
	}

	results := make([]SearchResult, len(list))
	for i, doc := range list {
		results[i] = SearchResult{
			Doc:   doc,
			Score: 1.0,
		}
	}
	return results
}

// Count returns total indexed documents.
func (e *Engine) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.docs)
}

// SourceCounts returns document distribution grouped by source.
func (e *Engine) SourceCounts() map[document.SourceType]int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	counts := make(map[document.SourceType]int)
	for _, doc := range e.docs {
		counts[doc.Source]++
	}
	return counts
}
