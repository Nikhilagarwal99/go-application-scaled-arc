package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/cache"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/database"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *cache.Client
}

func NewHealthHandler(db *gorm.DB, redis *cache.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Check runs all dependency pings concurrently and returns
// 200 if everything is healthy, 503 if anything is down.
func (h *HealthHandler) Check(c *gin.Context) {
	// Run both pings concurrently — no point waiting for postgres
	// to finish before checking redis, they are independent
	type result struct {
		name string
		err  error
	}

	results := make(chan result, 2)

	go func() {
		err := database.Ping(c.Request.Context(), h.db)
		results <- result{name: "postgres", err: err}
	}()

	go func() {
		err := h.redis.Ping(c.Request.Context())
		results <- result{name: "redis", err: err}
	}()

	// Collect both results
	services := make(map[string]string, 2)
	healthy := true

	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil {
			services[r.name] = r.err.Error()
			healthy = false
		} else {
			services[r.name] = "ok"
		}
	}

	status := "ok"
	httpStatus := http.StatusOK

	if !healthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable // 503
	}

	c.JSON(httpStatus, gin.H{
		"status":   status,
		"services": services,
	})
}
