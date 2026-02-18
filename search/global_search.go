package search

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/sahilm/fuzzy"
)

// Result represents a single search result.
type Result struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Subtitle     string  `json:"subtitle,omitempty"`
	URL          string  `json:"url"`
	Icon         string  `json:"icon,omitempty"`
	ResourceType string  `json:"resource_type"`
	Score        float64 `json:"score"`
}

// Searchable is the interface for resources that support global search.
type Searchable interface {
	// GetSearchableFields returns the fields to search in.
	GetSearchableFields() []string
	// Search performs a search and returns results.
	Search(ctx context.Context, query string, limit int) ([]Result, error)
	// GetSearchLabel returns the display label for this resource type.
	GetSearchLabel() string
	// GetSearchIcon returns the icon for this resource type.
	GetSearchIcon() string
	// GetSearchPriority returns the priority (lower = higher priority).
	GetSearchPriority() int
	// IsSearchEnabled returns whether search is enabled for this resource.
	IsSearchEnabled() bool
}

// BaseSearchable provides default implementations for Searchable.
type BaseSearchable struct {
	label    string
	icon     string
	priority int
	enabled  bool
	fields   []string
	searcher func(ctx context.Context, query string, limit int) ([]Result, error)
}

// NewSearchable creates a new searchable resource.
func NewSearchable(label string) *BaseSearchable {
	return &BaseSearchable{
		label:    label,
		icon:     "search",
		priority: 100,
		enabled:  true,
		fields:   make([]string, 0),
	}
}

func (s *BaseSearchable) GetSearchLabel() string        { return s.label }
func (s *BaseSearchable) GetSearchIcon() string         { return s.icon }
func (s *BaseSearchable) GetSearchPriority() int        { return s.priority }
func (s *BaseSearchable) IsSearchEnabled() bool         { return s.enabled }
func (s *BaseSearchable) GetSearchableFields() []string { return s.fields }

func (s *BaseSearchable) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if s.searcher != nil {
		return s.searcher(ctx, query, limit)
	}
	return []Result{}, nil
}

// SetIcon sets the search icon.
func (s *BaseSearchable) SetIcon(icon string) *BaseSearchable {
	s.icon = icon
	return s
}

// SetPriority sets the search priority.
func (s *BaseSearchable) SetPriority(priority int) *BaseSearchable {
	s.priority = priority
	return s
}

// SetEnabled sets whether search is enabled.
func (s *BaseSearchable) SetEnabled(enabled bool) *BaseSearchable {
	s.enabled = enabled
	return s
}

// SetFields sets the searchable fields.
func (s *BaseSearchable) SetFields(fields ...string) *BaseSearchable {
	s.fields = fields
	return s
}

// WithSearcher sets the search function.
func (s *BaseSearchable) WithSearcher(fn func(ctx context.Context, query string, limit int) ([]Result, error)) *BaseSearchable {
	s.searcher = fn
	return s
}

// Registry manages searchable resources.
type Registry struct {
	mu          sync.RWMutex
	searchables []Searchable
}

var globalRegistry = &Registry{
	searchables: make([]Searchable, 0),
}

// Register registers a searchable resource.
func Register(s Searchable) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.searchables = append(globalRegistry.searchables, s)
}

// Unregister removes a searchable by label.
func Unregister(label string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	filtered := make([]Searchable, 0)
	for _, s := range globalRegistry.searchables {
		if s.GetSearchLabel() != label {
			filtered = append(filtered, s)
		}
	}
	globalRegistry.searchables = filtered
}

// GetSearchables returns all registered searchables sorted by priority.
func GetSearchables() []Searchable {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	sorted := make([]Searchable, len(globalRegistry.searchables))
	copy(sorted, globalRegistry.searchables)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetSearchPriority() < sorted[j].GetSearchPriority()
	})

	return sorted
}

// SearchOptions configures a global search.
type SearchOptions struct {
	Query    string
	Limit    int
	Types    []string // Filter by resource types (empty = all)
	MinScore float64  // Minimum score threshold
}

// DefaultSearchOptions returns default search options.
func DefaultSearchOptions(query string) *SearchOptions {
	return &SearchOptions{
		Query:    query,
		Limit:    20,
		Types:    nil,
		MinScore: 0,
	}
}

// GlobalSearch performs a search across all registered searchables.
func GlobalSearch(ctx context.Context, opts *SearchOptions) ([]Result, error) {
	searchables := GetSearchables()

	if len(searchables) == 0 {
		return []Result{}, nil
	}

	// Calculate per-resource limit
	perResourceLimit := opts.Limit / len(searchables)
	if perResourceLimit < 3 {
		perResourceLimit = 3
	}

	var (
		allResults []Result
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	for _, s := range searchables {
		if !s.IsSearchEnabled() {
			continue
		}

		// Filter by types if specified
		if len(opts.Types) > 0 {
			found := false
			for _, t := range opts.Types {
				if strings.EqualFold(t, s.GetSearchLabel()) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		wg.Add(1)
		go func(searchable Searchable) {
			defer wg.Done()

			results, err := searchable.Search(ctx, opts.Query, perResourceLimit)
			if err != nil {
				return
			}

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(s)
	}

	wg.Wait()

	// Filter by minimum score
	if opts.MinScore > 0 {
		filtered := make([]Result, 0)
		for _, r := range allResults {
			if r.Score >= opts.MinScore {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	}

	// Sort by score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Limit total results
	if len(allResults) > opts.Limit {
		allResults = allResults[:opts.Limit]
	}

	return allResults, nil
}

// QuickSearch performs a quick search with default options.
func QuickSearch(ctx context.Context, query string) ([]Result, error) {
	return GlobalSearch(ctx, DefaultSearchOptions(query))
}

// SearchByType performs a search filtered by resource type.
func SearchByType(ctx context.Context, query string, resourceType string, limit int) ([]Result, error) {
	return GlobalSearch(ctx, &SearchOptions{
		Query: query,
		Limit: limit,
		Types: []string{resourceType},
	})
}

// CalculateScore calculates a relevance score using fuzzy matching.
// Returns a value between 0.0 and 1.0. Uses sahilm/fuzzy for scoring,
// with fallback to substring matching for exact/prefix hits.
func CalculateScore(query, text string) float64 {
	if query == "" || text == "" {
		return 0
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	// Exact match
	if lowerText == lowerQuery {
		return 1.0
	}
	// Prefix match
	if strings.HasPrefix(lowerText, lowerQuery) {
		return 0.95
	}
	// Substring match
	if strings.Contains(lowerText, lowerQuery) {
		return 0.8
	}

	// Fuzzy match via sahilm/fuzzy
	matches := fuzzy.Find(query, []string{text})
	if len(matches) > 0 {
		// fuzzy.Match.Score is negative (lower = worse); normalise to 0..0.75
		score := matches[0].Score
		if score >= 0 {
			return 0.75
		}
		// Map negative scores: -1 -> 0.74, -100 -> ~0.0
		normalized := 0.75 + float64(score)/200.0
		if normalized < 0 {
			normalized = 0
		}
		return normalized
	}

	return 0
}

// HighlightMatch highlights matching text in a result.
func HighlightMatch(text, query string) string {
	if query == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerText, lowerQuery)
	if idx == -1 {
		return text
	}

	return text[:idx] + "<mark>" + text[idx:idx+len(query)] + "</mark>" + text[idx+len(query):]
}

// Clear removes all registered searchables.
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.searchables = make([]Searchable, 0)
}

// Count returns the number of registered searchables.
func Count() int {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return len(globalRegistry.searchables)
}
