package incident

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/portable-siem/siem/internal/models"
)

type Manager struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{pool: pool}
}

func (m *Manager) Create(ctx context.Context, inc *models.Incident) error {
	if inc.ID == uuid.Nil {
		inc.ID = uuid.New()
	}
	now := time.Now()
	inc.CreatedAt = now
	inc.UpdatedAt = now
	inc.Status = models.IncidentStatusOpen

	_, err := m.pool.Exec(ctx, `
		INSERT INTO incidents (id, created_at, updated_at, title, description, severity, status, assigned_to, alert_ids, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		inc.ID, inc.CreatedAt, inc.UpdatedAt, inc.Title, inc.Description,
		string(inc.Severity), string(inc.Status), inc.AssignedTo, inc.AlertIDs, inc.Tags,
	)
	return err
}

func (m *Manager) Get(ctx context.Context, id uuid.UUID) (*models.Incident, error) {
	inc := &models.Incident{}
	err := m.pool.QueryRow(ctx, `
		SELECT id,created_at,updated_at,title,description,severity,status,assigned_to,alert_ids,tags
		FROM incidents WHERE id=$1`, id).
		Scan(&inc.ID, &inc.CreatedAt, &inc.UpdatedAt, &inc.Title, &inc.Description,
			&inc.Severity, &inc.Status, &inc.AssignedTo, &inc.AlertIDs, &inc.Tags)
	if err != nil {
		return nil, err
	}

	rows, err := m.pool.Query(ctx, `SELECT id,incident_id,timestamp,author,type,content FROM timeline_entries WHERE incident_id=$1 ORDER BY timestamp`, id)
	if err != nil {
		return inc, nil
	}
	defer rows.Close()
	for rows.Next() {
		entry := models.TimelineEntry{}
		_ = rows.Scan(&entry.ID, &entry.IncidentID, &entry.Timestamp, &entry.Author, &entry.Type, &entry.Content)
		inc.Timeline = append(inc.Timeline, entry)
	}
	return inc, nil
}

func (m *Manager) List(ctx context.Context, status models.IncidentStatus, limit, offset int) ([]*models.Incident, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where = "status=$1"
		args = append(args, string(status))
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := m.pool.Query(ctx, fmt.Sprintf(
		"SELECT id,created_at,updated_at,title,description,severity,status,assigned_to,alert_ids,tags FROM incidents WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d",
		where, limit, offset), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		inc := &models.Incident{}
		_ = rows.Scan(&inc.ID, &inc.CreatedAt, &inc.UpdatedAt, &inc.Title, &inc.Description,
			&inc.Severity, &inc.Status, &inc.AssignedTo, &inc.AlertIDs, &inc.Tags)
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

func (m *Manager) UpdateStatus(ctx context.Context, id uuid.UUID, status models.IncidentStatus) error {
	_, err := m.pool.Exec(ctx, `UPDATE incidents SET status=$1, updated_at=NOW() WHERE id=$2`, string(status), id)
	return err
}

func (m *Manager) AddTimelineEntry(ctx context.Context, entry *models.TimelineEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	entry.Timestamp = time.Now()

	_, err := m.pool.Exec(ctx, `
		INSERT INTO timeline_entries (id, incident_id, timestamp, author, type, content)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		entry.ID, entry.IncidentID, entry.Timestamp, entry.Author, entry.Type, entry.Content,
	)
	return err
}
