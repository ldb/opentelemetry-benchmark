package command

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

// Log is an HTTP middleware that logs all requests to the default log.Logger.
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("%s: %s %s", request.RemoteAddr, request.Method, request.URL.Path)
		handler.ServeHTTP(writer, request)
	})
}

// compressionWriter is an HTTP middleware that transparently enables Gzip compression on the response.
type compressionWriter struct {
	io.Writer
	http.ResponseWriter
}

func (c compressionWriter) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

func Gzip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(writer, request)
			return
		}
		writer.Header().Set("Content-Encoding", "gzip")
		gz, _ := gzip.NewWriterLevel(writer, gzip.BestCompression) // Ignore error, because Level is valid.
		defer gz.Close()
		gzw := compressionWriter{Writer: gz, ResponseWriter: writer}
		handler.ServeHTTP(gzw, request)
	})
}
