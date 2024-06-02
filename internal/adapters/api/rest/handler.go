package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/short-link/internal/adapters/database"
	"github.com/playmixer/short-link/internal/adapters/models"
	"go.uber.org/zap"
)

func (s *Server) handlerMain(c *gin.Context) {
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	c.Writer.Header().Add(ContentType, "text/plain")

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("can`t read body from request", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		err := c.Request.Body.Close()
		if err != nil {
			s.log.Error("failed close body request", zap.Error(err))
		}
	}()

	link := strings.TrimSpace(string(b))
	_, err = url.ParseRequestURI(link)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink, err := s.short.Shorty(ctx, link)
	if err != nil {
		s.log.Error(fmt.Sprintf("can`t shorted URI `%s`", b), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, fmt.Sprintf("%s/%s", s.baseURL, sLink))
}

func (s *Server) handlerShort(c *gin.Context) {
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	c.Writer.Header().Add(ContentType, "text/plain")

	id := c.Param("id")
	if id == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	link, err := s.short.GetURL(ctx, id)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c.Writer.Header().Add("Location", link)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) handlerAPIShorten(c *gin.Context) {
	ctx := context.Background()
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("can`t read body from request", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = c.Request.Body.Close() }()

	var req struct {
		URL string `json:"url"`
	}

	err = json.Unmarshal(b, &req)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = url.ParseRequestURI(req.URL)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink, err := s.short.Shorty(ctx, req.URL)
	if err != nil {
		s.log.Error(fmt.Sprintf("can`t shorted URI `%s`", b), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Add(ContentType, "application/json")
	c.JSON(http.StatusCreated, gin.H{
		"result": fmt.Sprintf("%s/%s", s.baseURL, sLink),
	})
}

func (s *Server) handlerPing(c *gin.Context) {
	conn, err := database.Conn()
	if err != nil {
		s.log.Info("failed create connect to database", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = conn.Ping()
	if err != nil {
		s.log.Info("failed ping database", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func (s *Server) handlerAPIShortenBatch(c *gin.Context) {
	ctx := context.Background()
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("can`t read body from request", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = c.Request.Body.Close() }()

	var req []models.ShortenBatchRequest
	err = json.Unmarshal(b, &req)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, v := range req {
		_, err = url.ParseRequestURI(v.OriginalURL)
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	sLink, err := s.short.ShortyBatch(ctx, req)
	if err != nil {
		s.log.Error(fmt.Sprintf("can`t shorted URI `%s`", b), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	for i := range sLink {
		sLink[i].ShortURL = fmt.Sprintf("%s/%s", s.baseURL, sLink[i].ShortURL)
	}

	c.Writer.Header().Add(ContentType, "application/json")
	c.JSON(http.StatusCreated, sLink)
}
