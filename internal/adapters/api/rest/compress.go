package rest

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

// GzipReader оборачивает io.ReadCloser и имплементирует интерфейс io.ReadCloser.
type GzipReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

// NewGzipReader создает GzipReader.
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

// Read читает из GzipReader.
func (gr GzipReader) Read(p []byte) (n int, err error) {
	n, err = gr.gz.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, fmt.Errorf("failed reading []byte %w", err)
	}
	return
}

// Close закрывает.
func (gr *GzipReader) Close() (err error) {
	err1 := gr.gz.Close()
	err2 := gr.r.Close()
	err = errors.Join(err1, err2)
	if err != nil {
		return fmt.Errorf("failed close reader %w", err)
	}

	return
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// NewGzipWriter оборачивает gin.ResponseWriter.
func NewGzipWriter(c *gin.Context) *gzipWriter {
	return &gzipWriter{
		c.Writer,
		gzip.NewWriter(c.Writer),
	}
}

// Write записывает в тело ответа.
func (gw *gzipWriter) Write(p []byte) (n int, err error) {
	n, err = gw.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed write gzip: %w", err)
	}
	return
}

// WriteString записывает строку в тело ответа.
func (gw *gzipWriter) WriteString(s string) (n int, err error) {
	n, err = gw.writer.Write([]byte(s))
	if err != nil {
		return n, fmt.Errorf("failed write gzip from string: %w", err)
	}
	return
}
