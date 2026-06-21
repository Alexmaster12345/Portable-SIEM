package models

import "time"

type RuleType string

const (
	RuleTypeThreshold   RuleType = "threshold"
	RuleTypeSequence    RuleType = "sequence"
	RuleTypeAnomaly     RuleType = "anomaly"
	RuleTypeCorrelation RuleType = "correlation"
)

type RuleAction string

const (
	RuleActionAlert    RuleAction = "alert"
	RuleActionBlock    RuleAction = "block"
	RuleActionEnrich   RuleAction = "enrich"
	RuleActionEscalate RuleAction = "escalate"
)

type Rule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        RuleType          `json:"type"`
	Enabled     bool              `json:"enabled"`
	Severity    Severity          `json:"severity"`
	Source      string            `json:"source,omitempty"`
	EventType   string            `json:"event_type,omitempty"`
	Conditions  []RuleCondition   `json:"conditions"`
	Threshold   int               `json:"threshold,omitempty"`
	WindowSecs  int               `json:"window_secs,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Actions     []RuleAction      `json:"actions"`
	MitreIDs    []string          `json:"mitre_ids,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type RuleCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // eq, ne, gt, lt, contains, regex
	Value    string `json:"value"`
}
