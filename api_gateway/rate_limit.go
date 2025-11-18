package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	tokens   int
}

var (
	visitors   = make(map[string]*visitor)
	visitorsMu sync.Mutex
)

const (
	rateLimitWindow    = time.Second // окно 1 секунда
	rateLimitMaxTokens = 10          // 10 запросов/сек на IP
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		visitorsMu.Lock()
		v, ok := visitors[ip]
		if !ok || now.Sub(v.lastSeen) > rateLimitWindow {
			v = &visitor{
				lastSeen: now,
				tokens:   rateLimitMaxTokens,
			}
			visitors[ip] = v
		}

		if v.tokens <= 0 {
			visitorsMu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests, slow down",
				},
			})
			return
		}

		v.tokens--
		v.lastSeen = now
		visitorsMu.Unlock()

		c.Next()
	}
}
