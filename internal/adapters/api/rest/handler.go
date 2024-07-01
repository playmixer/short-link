package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
	"go.uber.org/zap"
)

func (s *Server) handlerMain(c *gin.Context) {
	ctx := c.Request.Context()

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

	userID, err := s.checkAuth(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sLink, err := s.short.Shorty(ctx, userID, link)
	if err != nil {
		if errors.Is(err, storeerror.ErrNotUnique) {
			c.String(http.StatusConflict, s.baseLink(sLink))
			return
		}
		s.log.Error("can't shorten URI", zap.String("URI", string(b)), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, s.baseLink(sLink))
}

func (s *Server) handlerShort(c *gin.Context) {
	ctx := c.Request.Context()

	c.Writer.Header().Add(ContentType, "text/plain")

	id := c.Param("id")
	if id == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	link, err := s.short.GetURL(ctx, id)
	if err != nil {
		if errors.Is(err, storeerror.ErrShortURLDeleted) {
			c.Writer.WriteHeader(http.StatusGone)
			return
		}
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c.Writer.Header().Add("Location", link)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) handlerAPIShorten(c *gin.Context) {
	ctx := c.Request.Context()
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

	userID, err := s.checkAuth(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sLink, err := s.short.Shorty(ctx, userID, req.URL)
	if err != nil {
		if errors.Is(err, storeerror.ErrNotUnique) {
			c.Writer.Header().Add(ContentType, ApplicationJSON)
			c.JSON(http.StatusConflict, gin.H{
				"result": s.baseLink(sLink),
			})
			return
		}
		s.log.Error(fmt.Sprintf("can`t shorted URI `%s`", b), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Add(ContentType, ApplicationJSON)
	c.JSON(http.StatusCreated, gin.H{
		"result": s.baseLink(sLink),
	})
}

func (s *Server) handlerPing(c *gin.Context) {
	err := s.short.PingStore(c.Request.Context())
	if err != nil {
		s.log.Info("failed ping database", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func (s *Server) handlerAPIShortenBatch(c *gin.Context) {
	ctx := c.Request.Context()
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

	userID, err := s.checkAuth(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sLink, err := s.short.ShortyBatch(ctx, userID, req)
	for i, v := range sLink {
		sLink[i].ShortURL = s.baseLink(v.ShortURL)
	}
	if err != nil {
		if errors.Is(err, storeerror.ErrNotUnique) {
			c.Writer.Header().Add(ContentType, ApplicationJSON)
			c.JSON(http.StatusConflict, sLink)
			return
		}
		s.log.Error("can`t shorted URI", zap.String("URI", string(b)), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Add(ContentType, ApplicationJSON)
	c.JSON(http.StatusCreated, sLink)
}

func (s *Server) handlerAPIGetUserURLs(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := s.checkAuth(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	links, err := s.short.GetAllURL(ctx, userID)
	if err != nil {
		s.log.Error("can`t getting URLs by user", zap.String(CookieNameUserID, userID), zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := range links {
		links[i].ShortURL = s.baseLink(links[i].ShortURL)
	}
	if len(links) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	c.Writer.Header().Add(ContentType, ApplicationJSON)
	c.JSON(http.StatusOK, links)
}

func (s *Server) handlerAPIDeleteUserURLs(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := s.checkAuth(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("failed read body from request", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jBody []string
	err = json.Unmarshal(body, &jBody)
	if err != nil {
		s.log.Debug("invalid body", zap.Error(err))
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data := []models.ShortLink{}
	for _, short := range jBody {
		data = append(data, models.ShortLink{UserID: userID, ShortURL: short})
	}

	err = s.short.DeleteShortURLs(ctx, data)
	if err != nil {
		s.log.Error("delete short url error", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusAccepted)
}
