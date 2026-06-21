package correlation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/rules"
	"github.com/portable-siem/siem/internal/storage"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

// Engine correlates events using threshold-based rules.
type Engine struct {
	ruleEngine *rules.Engine
	redis      *storage.RedisStore
	alerts     chan<- *models.Alert
}

func NewEngine(ruleEngine *rules.Engine, redis *storage.RedisStore, alerts chan<- *models.Alert) *Engine {
	return &Engine{ruleEngine: ruleEngine, redis: redis, alerts: alerts}
}

// Process evaluates an event against all correlation rules.
func (e *Engine) Process(ctx context.Context, event *models.Event) {
	matched := e.ruleEngine.Match(ctx, event)
	for _, rule := range matched {
		switch rule.Type {
		case models.RuleTypeThreshold:
			e.processThreshold(ctx, rule, event)
		case models.RuleTypeCorrelation:
			e.processThreshold(ctx, rule, event)
		}
	}
}

func (e *Engine) processThreshold(ctx context.Context, rule *models.Rule, event *models.Event) {
	groupKey := buildGroupKey(rule, event)
	redisKey := fmt.Sprintf("corr:%s:%s", rule.ID, groupKey)

	count, err := e.redis.IncrWindowCounter(ctx, redisKey, rule.WindowSecs)
	if err != nil {
		logger.Error("redis incr error", zap.Error(err))
		return
	}

	threshold := rule.Threshold
	if threshold <= 0 {
		threshold = 1
	}

	if count == int64(threshold) {
		_ = e.redis.ResetWindowCounter(ctx, redisKey)
		alert := buildAlert(rule, event, groupKey, count)
		select {
		case e.alerts <- alert:
		case <-ctx.Done():
		}
	}
}

func buildGroupKey(rule *models.Rule, event *models.Event) string {
	parts := make([]string, 0, len(rule.GroupBy))
	for _, field := range rule.GroupBy {
		val := resolveField(field, event)
		parts = append(parts, val)
	}
	if len(parts) == 0 {
		return event.Host
	}
	return strings.Join(parts, ":")
}

func buildAlert(rule *models.Rule, event *models.Event, groupKey string, count int64) *models.Alert {
	return &models.Alert{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Status:      models.AlertStatusOpen,
		Title:       fmt.Sprintf("%s - %s", rule.Name, groupKey),
		Description: fmt.Sprintf("Rule '%s' triggered: %d events matched in window (group: %s)", rule.Name, count, groupKey),
		Host:        event.Host,
		EventIDs:    []uuid.UUID{event.ID},
		MitreIDs:    rule.MitreIDs,
		Fields: map[string]string{
			"group_key": groupKey,
			"count":     fmt.Sprintf("%d", count),
			"rule_type": string(rule.Type),
		},
	}
}

func resolveField(field string, event *models.Event) string {
	switch field {
	case "host":
		return event.Host
	case "source":
		return event.Source
	case "event_type":
		return event.EventType
	default:
		return event.Fields[field]
	}
}

// ImpossibleTravelDetector maintains user login locations and detects geographic anomalies.
type ImpossibleTravelDetector struct {
	redis  *storage.RedisStore
	alerts chan<- *models.Alert
}

func NewImpossibleTravelDetector(redis *storage.RedisStore, alerts chan<- *models.Alert) *ImpossibleTravelDetector {
	return &ImpossibleTravelDetector{redis: redis, alerts: alerts}
}

func (d *ImpossibleTravelDetector) Process(ctx context.Context, event *models.Event) {
	if event.EventType != "login_success" {
		return
	}
	username := event.Fields["username"]
	country := event.Fields["country"] // set by GeoIP enricher
	if username == "" || country == "" {
		return
	}

	key := fmt.Sprintf("travel:%s", username)
	type lastLogin struct {
		Country   string    `json:"country"`
		Timestamp time.Time `json:"timestamp"`
	}

	var last lastLogin
	if err := d.redis.Get(ctx, key, &last); err == nil {
		if last.Country != country {
			elapsed := time.Since(last.Timestamp)
			if elapsed < 2*time.Hour {
				alert := &models.Alert{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					RuleID:    "impossible_travel",
					RuleName:  "Impossible Travel",
					Severity:  models.SeverityCritical,
					Status:    models.AlertStatusOpen,
					Title:     fmt.Sprintf("Impossible Travel: %s", username),
					Description: fmt.Sprintf(
						"User %s logged in from %s then %s within %.0f minutes",
						username, last.Country, country, elapsed.Minutes(),
					),
					Host:     event.Host,
					EventIDs: []uuid.UUID{event.ID},
					MitreIDs: []string{"T1078"},
					Fields: map[string]string{
						"username":      username,
						"prev_country":  last.Country,
						"curr_country":  country,
						"elapsed_mins":  fmt.Sprintf("%.0f", elapsed.Minutes()),
					},
				}
				select {
				case d.alerts <- alert:
				case <-ctx.Done():
				}
			}
		}
	}

	_ = d.redis.Set(ctx, key, lastLogin{Country: country, Timestamp: time.Now()}, 24*time.Hour)
}
