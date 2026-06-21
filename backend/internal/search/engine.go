package search

import (
	"context"
	"time"

	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/storage"
)

// Engine provides full-text and field-based search over events.
type Engine struct {
	store *storage.PostgresStore
}

func NewEngine(store *storage.PostgresStore) *Engine {
	return &Engine{store: store}
}

type Query struct {
	Text      string
	From      *time.Time
	To        *time.Time
	Host      string
	Source    string
	EventType string
	Severity  string
	Limit     int
	Offset    int
}

type Result struct {
	Events []*models.Event `json:"events"`
	Total  int64           `json:"total"`
	Took   int64           `json:"took_ms"`
}

func (e *Engine) Search(ctx context.Context, q Query) (*Result, error) {
	start := time.Now()

	filter := models.EventFilter{
		From:      q.From,
		To:        q.To,
		Host:      q.Host,
		Source:    q.Source,
		EventType: q.EventType,
		Severity:  models.Severity(q.Severity),
		Query:     q.Text,
		Limit:     q.Limit,
		Offset:    q.Offset,
	}

	events, total, err := e.store.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &Result{
		Events: events,
		Total:  total,
		Took:   time.Since(start).Milliseconds(),
	}, nil
}

// Aggregate returns event counts grouped by a field over a time range.
func (e *Engine) Stats(ctx context.Context, since time.Time) (*models.EventStats, error) {
	return e.store.EventStats(ctx, since)
}
