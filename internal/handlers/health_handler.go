package handlers

import (
	"context"
	"net/http"
	"time"

	"blindtag_api/internal/database"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db        *database.DB
	startTime time.Time
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

// Check handles GCP Cloud Run health checks and liveness probes.
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "disconnected / degraded (" + err.Error() + ")"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "blindtag_api",
		"database":  dbStatus,
		"uptime":    time.Since(h.startTime).String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
