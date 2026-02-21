package search_test

import (
	"context"
	"testing"

	"github.com/bozz33/sublimego/search"
)

func TestBaseSearchableDefaults(t *testing.T) {
	s := search.NewSearchable("Users")
	if s.GetSearchLabel() != "Users" {
		t.Errorf("expected label 'Users', got %q", s.GetSearchLabel())
	}
	if !s.IsSearchEnabled() {
		t.Error("expected search to be enabled by default")
	}
	if s.GetSearchIcon() == "" {
		t.Error("expected non-empty default icon")
	}
}

func TestBaseSearchableSetters(t *testing.T) {
	s := search.NewSearchable("Posts").
		SetIcon("article").
		SetPriority(10)

	if s.GetSearchIcon() != "article" {
		t.Errorf("expected icon 'article', got %q", s.GetSearchIcon())
	}
	if s.GetSearchPriority() != 10 {
		t.Errorf("expected priority 10, got %d", s.GetSearchPriority())
	}
}

func TestBaseSearchableSearchEmpty(t *testing.T) {
	s := search.NewSearchable("Empty")
	results, err := s.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results without searcher, got %d", len(results))
	}
}

func TestBaseSearchableWithSearcher(t *testing.T) {
	s := search.NewSearchable("Products").WithSearcher(func(_ context.Context, query string, limit int) ([]search.Result, error) {
		if query == "apple" {
			return []search.Result{
				{ID: "1", Title: "Apple iPhone", URL: "/products/1", ResourceType: "products"},
				{ID: "2", Title: "Apple MacBook", URL: "/products/2", ResourceType: "products"},
			}, nil
		}
		return []search.Result{}, nil
	})

	results, err := s.Search(context.Background(), "apple", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].Title != "Apple iPhone" {
		t.Errorf("expected 'Apple iPhone', got %q", results[0].Title)
	}
}

func TestRegistryRegisterAndSearch(t *testing.T) {
	// Use the global registry API
	searchable := search.NewSearchable("TestArticles").WithSearcher(func(_ context.Context, query string, _ int) ([]search.Result, error) {
		if query == "golang" {
			return []search.Result{
				{ID: "1", Title: "Go Programming", URL: "/articles/1", ResourceType: "TestArticles"},
			}, nil
		}
		return nil, nil
	})
	search.Register(searchable)
	defer search.Unregister("TestArticles")

	results, err := search.GlobalSearch(context.Background(), search.DefaultSearchOptions("golang"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestCalculateScore(t *testing.T) {
	// CalculateScore(query, text) — "golang" contains prefix "go"
	score := search.CalculateScore("go", "golang")
	if score <= 0 {
		t.Errorf("expected positive score for text='golang' query='go', got %f", score)
	}

	noScore := search.CalculateScore("go", "python")
	if noScore >= score {
		t.Errorf("expected 'golang' to score higher than 'python' for query 'go', got %f vs %f", score, noScore)
	}
}
