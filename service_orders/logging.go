package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ctxKeyRequestID = "requestId"

// достать requestId из контекста
func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// middleware: читает/создаёт X-Request-ID и кладёт в контекст
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			// строго по ТЗ X-Request-ID должен создавать шлюз,
			// но тут делаем fallback для локальных тестов
			reqID = uuid.NewString()
		}

		c.Set(ctxKeyRequestID, reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)

		c.Next()
	}
}

// middleware: логирование запросов с requestId
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID := getRequestID(c)
		latency := time.Since(start)

		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		log.Printf(
			"requestId=%s method=%s path=%s status=%d duration=%s",
			reqID, method, path, status, latency,
		)
	}
}
