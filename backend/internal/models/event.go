package models

import (
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Event struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	Timestamp time.Time         `json:"timestamp" db:"timestamp"`
	ReceivedAt time.Time        `json:"received_at" db:"received_at"`
	Host      string            `json:"host" db:"host"`
	Source    string            `json:"source" db:"source"`
	EventType string            `json:"event_type" db:"event_type"`
	Severity  Severity          `json:"severity" db:"severity"`
	Message   string            `json:"message" db:"message"`
	Raw       string            `json:"raw,omitempty" db:"raw"`
	Fields    map[string]string `json:"fields" db:"fields"`
	Tags      []string          `json:"tags" db:"tags"`
	MitreIDs  []string          `json:"mitre_ids,omitempty" db:"mitre_ids"`
}

type EventFilter struct {
	From      *time.Time
	To        *time.Time
	Host      string
	Source    string
	EventType string
	Severity  Severity
	Query     string
	Limit     int
	Offset    int
}

type EventStats struct {
	Total    int64            `json:"total"`
	BySeverity map[string]int64 `json:"by_severity"`
	BySource   map[string]int64 `json:"by_source"`
	ByHost     map[string]int64 `json:"by_host"`
}
