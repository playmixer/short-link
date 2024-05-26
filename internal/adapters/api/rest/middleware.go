package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/short-link/internal/adapters/logger"
	"go.uber.org/zap"
)

func (s *Server) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Log.Info(
			"Request information",
			zap.String("uri", c.Request.RequestURI),
			zap.Duration("duration", time.Since(start)),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

func (s *Server) Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		cntnt := c.Request.Header.Get("Content-Type")
		if !strings.Contains(cntnt, "application/json") && !strings.Contains(cntnt, "text/html") {
			c.Next()
			return
		}

		if ok := strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip"); ok {
			gr, err := NewGzipReader(c.Request.Body)
			if err != nil {
				logger.Log.Warn(
					"GZip read error",
					zap.Error(err),
				)
				c.Writer.WriteHeader(http.StatusBadRequest)
				c.Abort()
				return
			}
			c.Request.Body = gr
			defer func() { _ = gr.Close() }()
		}

		if ok := strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip"); ok {
			c.Writer.Header().Set("Content-Encoding", "gzip")

			cw := NewGzipWriter(c)
			c.Writer = cw

			defer func() {
				_ = cw.writer.Close()
				c.Header("Content-Length", strconv.Itoa(c.Writer.Size()))
			}()
		}
		c.Next()
	}
}
