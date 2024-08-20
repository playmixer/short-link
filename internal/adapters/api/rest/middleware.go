package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger middleware логирования.
func (s *Server) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		s.log.Info(
			"Request information",
			zap.String("uri", c.Request.RequestURI),
			zap.Duration("duration", time.Since(start)),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

// GzipDecompress middleware распаковка сжатых данных.
func (s *Server) GzipDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok := strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip"); ok {
			gr, err := NewGzipReader(c.Request.Body)
			if err != nil {
				c.Writer.WriteHeader(http.StatusBadRequest)
				c.Abort()
				return
			}
			c.Request.Body = gr
			defer func() {
				if err := gr.Close(); err != nil {
					s.log.Info("failed close gzip reader", zap.Error(err))
				}
			}()
		}
		c.Next()
	}
}

// GzipCompress middleware запаковывает данные.
func (s *Server) GzipCompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok := strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip"); ok {
			c.Writer.Header().Set("Content-Encoding", "gzip")

			cw := NewGzipWriter(c)
			c.Writer = cw

			defer func() {
				if err := cw.writer.Close(); err != nil {
					s.log.Info("failed close gzip writer", zap.Error(err))
				}
			}()
		}
		c.Next()
	}
}

// CheckCookies middleware проверка куки файлов.
func (s *Server) CheckCookies() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ok bool
		var userCookie *http.Cookie
		userCookie, err := c.Request.Cookie(CookieNameUserID)
		if err == nil {
			_, ok = s.verifyJWT(userCookie.Value)
		}
		if err != nil || !ok {
			uniqueID := strconv.Itoa(time.Now().Nanosecond())
			signedCookie, err := s.CreateJWT(uniqueID)
			if err != nil {
				s.log.Info("failed sign cookies", zap.Error(err))
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Abort()
				return
			}
			userCookie = &http.Cookie{
				Name:  CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			}
			c.Request.AddCookie(userCookie)
		}

		http.SetCookie(c.Writer, userCookie)
		c.Next()
	}
}

// Auth middleware проверка аутентификации пользователя.
func (s *Server) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := s.checkAuth(c)
		if err != nil {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Abort()
		}

		c.Next()
	}
}
