package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/portable-siem/siem/internal/models"
	"github.com/portable-siem/siem/internal/rules"
)

type RulesHandler struct {
	engine *rules.Engine
}

func NewRulesHandler(engine *rules.Engine) *RulesHandler {
	return &RulesHandler{engine: engine}
}

func (h *RulesHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"rules": h.engine.ListRules()})
}

func (h *RulesHandler) Create(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	h.engine.AddRule(&rule)
	c.JSON(http.StatusCreated, rule)
}

func (h *RulesHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.ID = id
	rule.UpdatedAt = time.Now()
	h.engine.AddRule(&rule)
	c.JSON(http.StatusOK, rule)
}

func (h *RulesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.engine.RemoveRule(id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
