package alert

import (
	"context"

	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/storage"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

type Notifier interface {
	Send(ctx context.Context, alert *models.Alert) error
	Name() string
}

// Engine persists alerts and dispatches notifications.
type Engine struct {
	store     *storage.PostgresStore
	redis     *storage.RedisStore
	notifiers []Notifier
	in        <-chan *models.Alert
}

func NewEngine(store *storage.PostgresStore, redis *storage.RedisStore, in <-chan *models.Alert) *Engine {
	return &Engine{store: store, redis: redis, in: in}
}

func (e *Engine) AddNotifier(n Notifier) {
	e.notifiers = append(e.notifiers, n)
}

func (e *Engine) Run(ctx context.Context) {
	for {
		select {
		case alert, ok := <-e.in:
			if !ok {
				return
			}
			e.handle(ctx, alert)
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) handle(ctx context.Context, alert *models.Alert) {
	if err := e.store.InsertAlert(ctx, alert); err != nil {
		logger.Error("insert alert failed", zap.Error(err), zap.String("rule", alert.RuleName))
		return
	}

	// Broadcast to dashboard via Redis pub/sub
	_ = e.redis.Publish(ctx, "alerts", alert)

	for _, n := range e.notifiers {
		if err := n.Send(ctx, alert); err != nil {
			logger.Warn("notifier failed", zap.String("notifier", n.Name()), zap.Error(err))
		}
	}

	logger.Info("alert fired",
		zap.String("rule", alert.RuleName),
		zap.String("severity", string(alert.Severity)),
		zap.String("title", alert.Title),
	)
}
