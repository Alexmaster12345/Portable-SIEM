package rules

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

// Engine evaluates events against loaded rules.
type Engine struct {
	mu    sync.RWMutex
	rules map[string]*models.Rule
}

func NewEngine() *Engine {
	return &Engine{rules: make(map[string]*models.Rule)}
}

// LoadDir loads all JSON rule files from a directory.
func (e *Engine) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if err := e.LoadFile(filepath.Join(dir, entry.Name())); err != nil {
			logger.Warn("failed to load rule", zap.String("file", entry.Name()), zap.Error(err))
		}
	}
	return nil
}

func (e *Engine) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var rule models.Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return err
	}
	rule.UpdatedAt = time.Now()
	e.mu.Lock()
	e.rules[rule.ID] = &rule
	e.mu.Unlock()
	logger.Info("rule loaded", zap.String("id", rule.ID), zap.String("name", rule.Name))
	return nil
}

func (e *Engine) AddRule(rule *models.Rule) {
	e.mu.Lock()
	e.rules[rule.ID] = rule
	e.mu.Unlock()
}

func (e *Engine) RemoveRule(id string) {
	e.mu.Lock()
	delete(e.rules, id)
	e.mu.Unlock()
}

func (e *Engine) ListRules() []*models.Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*models.Rule, 0, len(e.rules))
	for _, r := range e.rules {
		out = append(out, r)
	}
	return out
}

// Match returns all rules that match the event (for threshold/sequence types,
// this only checks the filter conditions — the correlation engine handles counting).
func (e *Engine) Match(ctx context.Context, event *models.Event) []*models.Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var matched []*models.Rule
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		if matchesRule(rule, event) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func matchesRule(rule *models.Rule, event *models.Event) bool {
	if rule.Source != "" && rule.Source != event.Source {
		return false
	}
	if rule.EventType != "" && rule.EventType != event.EventType {
		return false
	}

	for _, cond := range rule.Conditions {
		val := resolveField(cond.Field, event)
		if !evalCondition(cond.Operator, val, cond.Value) {
			return false
		}
	}
	return true
}

func resolveField(field string, event *models.Event) string {
	switch field {
	case "host":
		return event.Host
	case "source":
		return event.Source
	case "event_type":
		return event.EventType
	case "severity":
		return string(event.Severity)
	case "message":
		return event.Message
	default:
		return event.Fields[field]
	}
}

func evalCondition(op, fieldVal, ruleVal string) bool {
	switch op {
	case "eq", "==":
		return fieldVal == ruleVal
	case "ne", "!=":
		return fieldVal != ruleVal
	case "contains":
		return strings.Contains(fieldVal, ruleVal)
	case "not_contains":
		return !strings.Contains(fieldVal, ruleVal)
	case "regex":
		matched, _ := regexp.MatchString(ruleVal, fieldVal)
		return matched
	case "gt":
		a, _ := strconv.ParseFloat(fieldVal, 64)
		b, _ := strconv.ParseFloat(ruleVal, 64)
		return a > b
	case "lt":
		a, _ := strconv.ParseFloat(fieldVal, 64)
		b, _ := strconv.ParseFloat(ruleVal, 64)
		return a < b
	}
	return false
}
