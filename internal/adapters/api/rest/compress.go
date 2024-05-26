package rest

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

type GzipReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	gzbody, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed reading request body %w", err)
	}
	defer func() { _ = r.Close() }()

	buf := bytes.NewReader(gzbody)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed reading gzip body %w", err)
	}
	return &GzipReader{
		r:  r,
		gz: zr,
	}, nil
}

func (gr GzipReader) Read(p []byte) (n int, err error) {
	n, err = gr.gz.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, fmt.Errorf("failed reading []byte %w", err)
	}
	return
}

func (gr *GzipReader) Close() (err error) {
	if err := gr.r.Close(); err != nil {
		return fmt.Errorf("failed closing reader %w", err)
	}
	err = gr.gz.Close()
	if err != nil {
		return fmt.Errorf("failed closing gzip reader %w", err)
	}
	return
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func NewGzipWriter(c *gin.Context) *gzipWriter {
	return &gzipWriter{
		c.Writer,
		gzip.NewWriter(c.Writer),
	}
}

func (gw *gzipWriter) Write(p []byte) (n int, err error) {
	gw.Header().Del(ContentLength)
	n, err = gw.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed write gzip: %w", err)
	}
	return
}

func (gw *gzipWriter) WriteString(s string) (n int, err error) {
	gw.Header().Del(ContentLength)
	n, err = gw.writer.Write([]byte(s))
	if err != nil {
		return n, fmt.Errorf("failed write gzip from string: %w", err)
	}
	return
}
