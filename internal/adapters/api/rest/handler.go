package rest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (s *Server) mainHandle(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("can`t read body from request, error: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = url.ParseRequestURI(string(b))
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink, err := s.short.Shorty(string(b))
	if err != nil {
		log.Printf("can`t shorted URI `%s`, error: %s", b, err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, fmt.Sprintf("%s/%s", s.baseURL, sLink))
}

func (s *Server) shortHandle(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")

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
