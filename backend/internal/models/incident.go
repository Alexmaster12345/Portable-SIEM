package models

import (
	"time"

	"github.com/google/uuid"
)

type IncidentStatus string
type IncidentSeverity string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
)

type Incident struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	Title       string         `json:"title" db:"title"`
	Description string         `json:"description" db:"description"`
	Severity    Severity       `json:"severity" db:"severity"`
	Status      IncidentStatus `json:"status" db:"status"`
	AssignedTo  string         `json:"assigned_to,omitempty" db:"assigned_to"`
	AlertIDs    []uuid.UUID    `json:"alert_ids" db:"alert_ids"`
	Tags        []string       `json:"tags" db:"tags"`
	Timeline    []TimelineEntry `json:"timeline"`
}

type TimelineEntry struct {
	ID         uuid.UUID `json:"id" db:"id"`
	IncidentID uuid.UUID `json:"incident_id" db:"incident_id"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Author     string    `json:"author" db:"author"`
	Type       string    `json:"type" db:"type"` // note, evidence, status_change, alert
	Content    string    `json:"content" db:"content"`
}

type IOC struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	Type       string    `json:"type" db:"type"` // ip, domain, hash, url
	Value      string    `json:"value" db:"value"`
	Confidence int       `json:"confidence" db:"confidence"` // 0-100
	Source     string    `json:"source" db:"source"`
	Tags       []string  `json:"tags" db:"tags"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}
