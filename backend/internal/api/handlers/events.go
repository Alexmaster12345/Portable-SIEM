package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/search"
	"github.com/portable-siem/siem/internal/storage"
)

type EventsHandler struct {
	store  *storage.PostgresStore
	search *search.Engine
}

func NewEventsHandler(store *storage.PostgresStore, search *search.Engine) *EventsHandler {
	return &EventsHandler{store: store, search: search}
}

func (h *EventsHandler) List(c *gin.Context) {
	filter := models.EventFilter{
		Host:      c.Query("host"),
		Source:    c.Query("source"),
		EventType: c.Query("event_type"),
		Severity:  models.Severity(c.Query("severity")),
		Query:     c.Query("q"),
	}

	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		}
	}

	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "100"))
	filter.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	events, total, err := h.store.QueryEvents(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *EventsHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	filter := models.EventFilter{Query: id.String(), Limit: 1}
	events, _, err := h.store.QueryEvents(c.Request.Context(), filter)
	if err != nil || len(events) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, events[0])
}

func (h *EventsHandler) Ingest(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if err := h.store.InsertEvent(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *EventsHandler) Stats(c *gin.Context) {
	since := time.Now().Add(-24 * time.Hour)
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = t
		}
	}
	stats, err := h.search.Stats(c.Request.Context(), since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
