package intelligence

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/storage"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

type Feed struct {
	Name   string
	URL    string
	Type   string // ip, domain, hash
	Source string
}

// Manager downloads threat feeds and stores IOCs.
type Manager struct {
	store  *storage.PostgresStore
	client *http.Client
	feeds  []Feed
}

func NewManager(store *storage.PostgresStore) *Manager {
	return &Manager{
		store:  store,
		client: &http.Client{Timeout: 30 * time.Second},
		feeds:  defaultFeeds(),
	}
}

func defaultFeeds() []Feed {
	return []Feed{
		{
			Name:   "Feodo Tracker",
			URL:    "https://feodotracker.abuse.ch/downloads/ipblocklist.txt",
			Type:   "ip",
			Source: "feodotracker",
		},
		{
			Name:   "URLhaus",
			URL:    "https://urlhaus.abuse.ch/downloads/text/",
			Type:   "url",
			Source: "urlhaus",
		},
	}
}

func (m *Manager) AddFeed(f Feed) {
	m.feeds = append(m.feeds, f)
}

// Refresh downloads all feeds and upserts IOCs into the database.
func (m *Manager) Refresh(ctx context.Context) error {
	for _, feed := range m.feeds {
		logger.Info("refreshing threat feed", zap.String("name", feed.Name))
		if err := m.fetchFeed(ctx, feed); err != nil {
			logger.Warn("feed refresh failed", zap.String("name", feed.Name), zap.Error(err))
		}
	}
	return nil
}

func (m *Manager) fetchFeed(ctx context.Context, feed Feed) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ioc := &models.IOC{
			Type:       feed.Type,
			Value:      line,
			Confidence: 80,
			Source:     feed.Source,
		}
		if err := m.store.UpsertIOC(ctx, ioc); err != nil {
			continue
		}
		count++
	}
	logger.Info("feed loaded", zap.String("name", feed.Name), zap.Int("count", count))
	return nil
}

// LoadFile imports IOCs from a local JSON file.
func (m *Manager) LoadFile(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ioc file: %w", err)
	}

	var iocs []*models.IOC
	if err := json.Unmarshal(data, &iocs); err != nil {
		return fmt.Errorf("parse ioc file: %w", err)
	}

	for _, ioc := range iocs {
		if err := m.store.UpsertIOC(ctx, ioc); err != nil {
			logger.Warn("upsert ioc failed", zap.String("value", ioc.Value), zap.Error(err))
		}
	}
	logger.Info("ioc file loaded", zap.String("path", path), zap.Int("count", len(iocs)))
	return nil
}

// StartAutoRefresh refreshes feeds on a schedule.
func (m *Manager) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = m.Refresh(ctx)
		case <-ctx.Done():
			return
		}
	}
}
