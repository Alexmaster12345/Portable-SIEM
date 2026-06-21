package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/portable-siem/siem/internal/models"
)

type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(ctx context.Context, alert *models.Alert) error {
	if s.webhookURL == "" {
		return nil
	}

	color := severityColor(alert.Severity)
	payload := map[string]any{
		"attachments": []map[string]any{
			{
				"color":   color,
				"title":   fmt.Sprintf("[%s] %s", alert.Severity, alert.Title),
				"text":    alert.Description,
				"footer":  "Portable SIEM",
				"ts":      alert.CreatedAt.Unix(),
				"fields": []map[string]string{
					{"title": "Rule", "value": alert.RuleName, "short": "true"},
					{"title": "Host", "value": alert.Host, "short": "true"},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

func severityColor(s models.Severity) string {
	switch s {
	case models.SeverityCritical:
		return "#FF0000"
	case models.SeverityHigh:
		return "#FF6600"
	case models.SeverityMedium:
		return "#FFAA00"
	default:
		return "#00AA00"
	}
}
