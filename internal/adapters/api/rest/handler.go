package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) mainHandle(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")

	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
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
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c.String(http.StatusCreated, fmt.Sprintf("%s/%s", s.baseURL, sLink))
}

func (s *Server) shortHandle(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")

	id := c.Request.URL.Path
	id = strings.ReplaceAll(id, "/", "")

	if id == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		s.log.ERROR(fmt.Sprintf("page `%s` not found", id))
		return
	}
	url, err := s.short.GetURL(id)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		s.log.ERROR(fmt.Sprintf("page not found by id `%s`, err: %e", id, err))
		return
	}
	c.Writer.Header().Add("Location", url)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}
