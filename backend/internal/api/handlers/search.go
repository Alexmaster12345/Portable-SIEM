package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/portable-siem/siem/internal/search"
)

type SearchHandler struct {
	engine *search.Engine
}

func NewSearchHandler(engine *search.Engine) *SearchHandler {
	return &SearchHandler{engine: engine}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := search.Query{
		Text:      c.Query("q"),
		Host:      c.Query("host"),
		Source:    c.Query("source"),
		EventType: c.Query("event_type"),
		Severity:  c.Query("severity"),
	}
	q.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "100"))
	q.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.From = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.To = &t
		}
	}

	result, err := h.engine.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
