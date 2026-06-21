package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/pkg/config"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(cfg config.DatabaseConfig) (*PostgresStore, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password, cfg.SSLMode,
	)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) Pool() *pgxpool.Pool { return s.pool }

// ---- Events ----

func (s *PostgresStore) InsertEvent(ctx context.Context, e *models.Event) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now()
	}

	fields, _ := json.Marshal(e.Fields)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO events (id, timestamp, received_at, host, source, event_type, severity, message, raw, fields, tags, mitre_ids)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		e.ID, e.Timestamp, e.ReceivedAt, e.Host, e.Source, e.EventType,
		string(e.Severity), e.Message, e.Raw, fields, e.Tags, e.MitreIDs,
	)
	return err
}

func (s *PostgresStore) InsertEventsBatch(ctx context.Context, events []*models.Event) error {
	batch := &pgx.Batch{}
	for _, e := range events {
		if e.ID == uuid.Nil {
			e.ID = uuid.New()
		}
		if e.ReceivedAt.IsZero() {
			e.ReceivedAt = time.Now()
		}
		fields, _ := json.Marshal(e.Fields)
		batch.Queue(`
			INSERT INTO events (id, timestamp, received_at, host, source, event_type, severity, message, raw, fields, tags, mitre_ids)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			e.ID, e.Timestamp, e.ReceivedAt, e.Host, e.Source, e.EventType,
			string(e.Severity), e.Message, e.Raw, fields, e.Tags, e.MitreIDs,
		)
	}
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range events {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) QueryEvents(ctx context.Context, f models.EventFilter) ([]*models.Event, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	add := func(cond string, val any) {
		where = append(where, fmt.Sprintf(cond, i))
		args = append(args, val)
		i++
	}

	if f.From != nil {
		add("timestamp >= $%d", f.From)
	}
	if f.To != nil {
		add("timestamp <= $%d", f.To)
	}
	if f.Host != "" {
		add("host = $%d", f.Host)
	}
	if f.Source != "" {
		add("source = $%d", f.Source)
	}
	if f.EventType != "" {
		add("event_type = $%d", f.EventType)
	}
	if f.Severity != "" {
		add("severity = $%d", string(f.Severity))
	}
	if f.Query != "" {
		add("to_tsvector('english', message) @@ plainto_tsquery('english', $%d)", f.Query)
	}

	clause := strings.Join(where, " AND ")

	var total int64
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM events WHERE "+clause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx,
		fmt.Sprintf("SELECT id,timestamp,received_at,host,source,event_type,severity,message,raw,fields,tags,mitre_ids FROM events WHERE %s ORDER BY timestamp DESC LIMIT %d OFFSET %d", clause, limit, f.Offset),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		e := &models.Event{}
		var rawFields []byte
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.ReceivedAt, &e.Host, &e.Source, &e.EventType, &e.Severity, &e.Message, &e.Raw, &rawFields, &e.Tags, &e.MitreIDs); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(rawFields, &e.Fields)
		events = append(events, e)
	}
	return events, total, nil
}

// ---- Alerts ----

func (s *PostgresStore) InsertAlert(ctx context.Context, a *models.Alert) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	fields, _ := json.Marshal(a.Fields)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO alerts (id,created_at,updated_at,rule_id,rule_name,severity,status,title,description,host,event_ids,mitre_ids,assigned_to,notes,incident_id,fields)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		a.ID, a.CreatedAt, a.UpdatedAt, a.RuleID, a.RuleName, string(a.Severity),
		string(a.Status), a.Title, a.Description, a.Host, a.EventIDs, a.MitreIDs,
		a.AssignedTo, a.Notes, a.IncidentID, fields,
	)
	return err
}

func (s *PostgresStore) ListAlerts(ctx context.Context, status models.AlertStatus, limit, offset int) ([]*models.Alert, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where = "status = $1"
		args = append(args, string(status))
	}

	var total int64
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alerts WHERE "+where, args...).Scan(&total)

	if limit <= 0 {
		limit = 50
	}
	query := fmt.Sprintf("SELECT id,created_at,updated_at,rule_id,rule_name,severity,status,title,description,host,event_ids,mitre_ids,assigned_to,notes,incident_id,fields FROM alerts WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d", where, limit, offset)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		a := &models.Alert{}
		var rawFields []byte
		if err := rows.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt, &a.RuleID, &a.RuleName, &a.Severity, &a.Status, &a.Title, &a.Description, &a.Host, &a.EventIDs, &a.MitreIDs, &a.AssignedTo, &a.Notes, &a.IncidentID, &rawFields); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(rawFields, &a.Fields)
		alerts = append(alerts, a)
	}
	return alerts, total, nil
}

func (s *PostgresStore) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status models.AlertStatus, notes string) error {
	_, err := s.pool.Exec(ctx, `UPDATE alerts SET status=$1, notes=$2, updated_at=NOW() WHERE id=$3`, string(status), notes, id)
	return err
}

// ---- IOC ----

func (s *PostgresStore) UpsertIOC(ctx context.Context, ioc *models.IOC) error {
	if ioc.ID == uuid.Nil {
		ioc.ID = uuid.New()
	}
	ioc.CreatedAt = time.Now()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO iocs (id,created_at,type,value,confidence,source,tags,expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (type,value) DO UPDATE SET confidence=EXCLUDED.confidence, source=EXCLUDED.source, expires_at=EXCLUDED.expires_at`,
		ioc.ID, ioc.CreatedAt, ioc.Type, ioc.Value, ioc.Confidence, ioc.Source, ioc.Tags, ioc.ExpiresAt,
	)
	return err
}

func (s *PostgresStore) LookupIOC(ctx context.Context, value string) (*models.IOC, error) {
	ioc := &models.IOC{}
	err := s.pool.QueryRow(ctx, `SELECT id,created_at,type,value,confidence,source,tags FROM iocs WHERE value=$1 AND (expires_at IS NULL OR expires_at > NOW()) LIMIT 1`, value).
		Scan(&ioc.ID, &ioc.CreatedAt, &ioc.Type, &ioc.Value, &ioc.Confidence, &ioc.Source, &ioc.Tags)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ioc, err
}

func (s *PostgresStore) EventStats(ctx context.Context, since time.Time) (*models.EventStats, error) {
	stats := &models.EventStats{
		BySeverity: map[string]int64{},
		BySource:   map[string]int64{},
		ByHost:     map[string]int64{},
	}

	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM events WHERE timestamp >= $1", since).Scan(&stats.Total)

	rows, _ := s.pool.Query(ctx, "SELECT severity, COUNT(*) FROM events WHERE timestamp >= $1 GROUP BY severity", since)
	for rows.Next() {
		var k string
		var v int64
		_ = rows.Scan(&k, &v)
		stats.BySeverity[k] = v
	}
	rows.Close()

	rows, _ = s.pool.Query(ctx, "SELECT source, COUNT(*) FROM events WHERE timestamp >= $1 GROUP BY source ORDER BY 2 DESC LIMIT 10", since)
	for rows.Next() {
		var k string
		var v int64
		_ = rows.Scan(&k, &v)
		stats.BySource[k] = v
	}
	rows.Close()

	rows, _ = s.pool.Query(ctx, "SELECT host, COUNT(*) FROM events WHERE timestamp >= $1 GROUP BY host ORDER BY 2 DESC LIMIT 10", since)
	for rows.Next() {
		var k string
		var v int64
		_ = rows.Scan(&k, &v)
		stats.ByHost[k] = v
	}
	rows.Close()

	return stats, nil
}
