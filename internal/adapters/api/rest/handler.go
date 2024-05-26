package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) handlerMain(c *gin.Context) {
	c.Writer.Header().Add(ContentType, "text/plain")

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("can`t read body from request, error: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = c.Request.Body.Close() }()

	link := strings.TrimSpace(string(b))
	_, err = url.ParseRequestURI(link)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink, err := s.short.Shorty(link)
	if err != nil {
		log.Printf("can`t shorted URI `%s`, error: %v", b, err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, fmt.Sprintf("%s/%s", s.baseURL, sLink))
}

func (s *Server) handlerShort(c *gin.Context) {
	c.Writer.Header().Add(ContentType, "text/plain")

	id := c.Param("id")
	if id == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	link, err := s.short.GetURL(id)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c.Writer.Header().Add("Location", link)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) handlerAPIShorten(c *gin.Context) {
	c.Writer.Header().Add(ContentType, "application/json")

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("can`t read body from request, error: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = c.Request.Body.Close() }()

	var req struct {
		URL string `json:"url"`
	}

	err = json.Unmarshal(b, &req)
	if err != nil {
		log.Printf("can`t unmarshal body from request, error: %v", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = url.ParseRequestURI(req.URL)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink, err := s.short.Shorty(req.URL)
	if err != nil {
		log.Printf("can`t shorted URI `%s`, error: %v", b, err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"result": fmt.Sprintf("%s/%s", s.baseURL, sLink),
	})
}
