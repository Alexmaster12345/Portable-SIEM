package models

import (
	"time"

	"github.com/google/uuid"
)

type AlertStatus string

const (
	AlertStatusOpen         AlertStatus = "open"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusFalsePositive AlertStatus = "false_positive"
)

type Alert struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	RuleID      string      `json:"rule_id" db:"rule_id"`
	RuleName    string      `json:"rule_name" db:"rule_name"`
	Severity    Severity    `json:"severity" db:"severity"`
	Status      AlertStatus `json:"status" db:"status"`
	Title       string      `json:"title" db:"title"`
	Description string      `json:"description" db:"description"`
	Host        string      `json:"host" db:"host"`
	EventIDs    []uuid.UUID `json:"event_ids" db:"event_ids"`
	MitreIDs    []string    `json:"mitre_ids" db:"mitre_ids"`
	AssignedTo  string      `json:"assigned_to,omitempty" db:"assigned_to"`
	Notes       string      `json:"notes,omitempty" db:"notes"`
	IncidentID  *uuid.UUID  `json:"incident_id,omitempty" db:"incident_id"`
	Fields      map[string]string `json:"fields" db:"fields"`
}
