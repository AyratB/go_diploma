package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	ResponseWriter   http.ResponseWriter
	CompressedWriter io.Writer
}

func (gw gzipWriter) Write(b []byte) (int, error) {
	return gw.CompressedWriter.Write(b)
}

func (gw gzipWriter) WriteHeader(statusCode int) {
	gw.ResponseWriter.WriteHeader(statusCode)
}

func (gw gzipWriter) Header() http.Header {
	return gw.ResponseWriter.Header()
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gzipReader
			defer gzipReader.Close()
		}

		writer := io.Writer(w)
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzWriter, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			defer gzWriter.Close()

			w.Header().Set("Content-Encoding", "gzip")
			writer = gzWriter
		}

		next.ServeHTTP(gzipWriter{ResponseWriter: w, CompressedWriter: writer}, r)
	})
}
