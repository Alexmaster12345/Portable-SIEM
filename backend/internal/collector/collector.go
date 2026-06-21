package collector

import (
	"context"
	"sync"

	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

// Collector is implemented by any log source.
type Collector interface {
	Name() string
	Start(ctx context.Context, out chan<- *models.Event) error
}

// Manager runs all registered collectors and fans their output into one channel.
type Manager struct {
	collectors []Collector
	out        chan *models.Event
}

func NewManager(bufferSize int) *Manager {
	return &Manager{
		out: make(chan *models.Event, bufferSize),
	}
}

func (m *Manager) Register(c Collector) {
	m.collectors = append(m.collectors, c)
}

// Events returns the unified output channel.
func (m *Manager) Events() <-chan *models.Event { return m.out }

func (m *Manager) Start(ctx context.Context) {
	var wg sync.WaitGroup
	for _, c := range m.collectors {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("collector started", zap.String("name", c.Name()))
			if err := c.Start(ctx, m.out); err != nil && ctx.Err() == nil {
				logger.Error("collector error", zap.String("name", c.Name()), zap.Error(err))
			}
		}()
	}
	go func() {
		wg.Wait()
		close(m.out)
	}()
}
