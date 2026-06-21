package api

import (
	"github.com/gin-gonic/gin"
	"github.com/portable-siem/siem/internal/alert"
	"github.com/portable-siem/siem/internal/api/handlers"
	"github.com/portable-siem/siem/internal/api/middleware"
	"github.com/portable-siem/siem/internal/incident"
	"github.com/portable-siem/siem/internal/rules"
	"github.com/portable-siem/siem/internal/search"
	"github.com/portable-siem/siem/internal/storage"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	store *storage.PostgresStore,
	redis *storage.RedisStore,
	searchEngine *search.Engine,
	alertEngine *alert.Engine,
	ruleEngine *rules.Engine,
	incidentMgr *incident.Manager,
) *Router {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "portable-siem"})
	})

	api := r.Group("/api/v1")

	// Events
	eventsHandler := handlers.NewEventsHandler(store, searchEngine)
	api.GET("/events", eventsHandler.List)
	api.GET("/events/:id", eventsHandler.Get)
	api.POST("/events", eventsHandler.Ingest)
	api.GET("/events/stats", eventsHandler.Stats)

	// Alerts
	alertsHandler := handlers.NewAlertsHandler(store)
	api.GET("/alerts", alertsHandler.List)
	api.GET("/alerts/:id", alertsHandler.Get)
	api.PATCH("/alerts/:id/status", alertsHandler.UpdateStatus)

	// Search
	searchHandler := handlers.NewSearchHandler(searchEngine)
	api.GET("/search", searchHandler.Search)

	// Rules
	rulesHandler := handlers.NewRulesHandler(ruleEngine)
	api.GET("/rules", rulesHandler.List)
	api.POST("/rules", rulesHandler.Create)
	api.PUT("/rules/:id", rulesHandler.Update)
	api.DELETE("/rules/:id", rulesHandler.Delete)

	// Incidents
	incidentsHandler := handlers.NewIncidentsHandler(incidentMgr)
	api.GET("/incidents", incidentsHandler.List)
	api.GET("/incidents/:id", incidentsHandler.Get)
	api.POST("/incidents", incidentsHandler.Create)
	api.PATCH("/incidents/:id/status", incidentsHandler.UpdateStatus)
	api.POST("/incidents/:id/timeline", incidentsHandler.AddTimelineEntry)

	return &Router{engine: r}
}

func (r *Router) Engine() *gin.Engine { return r.engine }
