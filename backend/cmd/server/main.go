package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/portable-siem/siem/internal/alert"
	"github.com/portable-siem/siem/internal/alert/notifier"
	"github.com/portable-siem/siem/internal/api"
	"github.com/portable-siem/siem/internal/collector"
	"github.com/portable-siem/siem/internal/collector/linux"
	"github.com/portable-siem/siem/internal/collector/network"
	"github.com/portable-siem/siem/internal/correlation"
	"github.com/portable-siem/siem/internal/incident"
	"github.com/portable-siem/siem/internal/intelligence"
	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/parser"
	"github.com/portable-siem/siem/internal/rules"
	"github.com/portable-siem/siem/internal/search"
	"github.com/portable-siem/siem/internal/storage"
	"github.com/portable-siem/siem/pkg/config"
	"github.com/portable-siem/siem/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfgPath := os.Getenv("SIEM_CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Server.Mode); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Portable SIEM starting")

	// Storage
	pgStore, err := storage.NewPostgresStore(cfg.Database)
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}
	defer pgStore.Close()

	redisStore, err := storage.NewRedisStore(cfg.Redis)
	if err != nil {
		logger.Fatal("redis connect failed", zap.Error(err))
	}
	defer redisStore.Close()

	// Engines
	ruleEngine := rules.NewEngine()
	if err := ruleEngine.LoadDir("rules/"); err != nil {
		logger.Warn("rules dir load failed", zap.Error(err))
	}

	alertCh := make(chan *models.Alert, 1000)

	alertEngine := alert.NewEngine(pgStore, redisStore, alertCh)
	if cfg.Alert.SlackWebhook != "" {
		alertEngine.AddNotifier(notifier.NewSlackNotifier(cfg.Alert.SlackWebhook))
	}

	correlationEngine := correlation.NewEngine(ruleEngine, redisStore, alertCh)
	travelDetector := correlation.NewImpossibleTravelDetector(redisStore, alertCh)

	// Enrichment pipeline
	pipeline := parser.NewPipeline(
		&parser.NormalizeEnricher{},
		parser.NewFieldExtractEnricher(),
		&parser.MITREEnricher{},
	)

	searchEngine := search.NewEngine(pgStore)
	incidentMgr := incident.NewManager(pgStore.Pool())

	// Collectors
	hostname, _ := os.Hostname()
	collectorMgr := collector.NewManager(10000)

	if cfg.Collectors.LinuxEnabled {
		collectorMgr.Register(linux.NewJournaldCollector(hostname, nil))
		collectorMgr.Register(linux.NewAuthLogCollector(hostname))
		collectorMgr.Register(linux.NewSyslogCollector(hostname))
	}
	collectorMgr.Register(network.NewSyslogReceiver(
		cfg.Collectors.SyslogAddress,
		cfg.Collectors.SyslogPort,
	))

	// Threat intelligence
	intelMgr := intelligence.NewManager(pgStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start everything
	collectorMgr.Start(ctx)
	go alertEngine.Run(ctx)
	go intelMgr.StartAutoRefresh(ctx, 6*time.Hour)

	// Event processing loop
	go func() {
		for event := range collectorMgr.Events() {
			event = pipeline.Process(event)

			if err := pgStore.InsertEvent(ctx, event); err != nil {
				logger.Error("insert event", zap.Error(err))
				continue
			}
			_ = redisStore.TrackHostLastSeen(ctx, event.Host)
			_ = redisStore.Publish(ctx, "events", event)

			correlationEngine.Process(ctx, event)
			travelDetector.Process(ctx, event)
		}
	}()

	// HTTP server
	router := api.NewRouter(pgStore, redisStore, searchEngine, alertEngine, ruleEngine, incidentMgr)
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router.Engine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Info("API server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	cancel()

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	_ = srv.Shutdown(shutCtx)

	logger.Info("goodbye")
}
