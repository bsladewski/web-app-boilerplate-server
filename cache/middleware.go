package cache

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LocalCacheMiddleware responds with values from the local cache if possible.
// If the request is not present in the cache add it to the cache with the
// specified time to live.
func LocalCacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {

		// only cache responses to GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// use the path and GET query as the cache key
		key := c.Request.URL.EscapedPath() + "?" + c.Request.URL.Query().Encode()

		// check if the request is cached, if so respond with the cached value
		if item, ok := GetLocal(key); ok {
			if resp, ok := item.(responseCacheItem); ok {
				c.Data(http.StatusOK, resp.ContentType, resp.Data)
				c.Abort()
				return
			}
		}

		// wrap the response writer so we can record the response to the
		// incoming request
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			responseData:   bytes.Buffer{},
		}
		c.Writer = writer

		// execute the next handler function
		c.Next()

		// cache the response
		SetLocal(key, responseCacheItem{
			ContentType: c.Writer.Header().Get("Content-Type"),
			Data:        writer.responseData.Bytes(),
		}, ttl)

	}
}

// responseCacheItem is used to store the raw response to an HTTP request in the
// local cache.
type responseCacheItem struct {
	ContentType string
	Data        []byte
}

// responseWriter is used to wrap the response writer used by the response cache
// middleware to simultaneously record and write the response so the raw
// response contents can be cached.
type responseWriter struct {
	gin.ResponseWriter
	responseData bytes.Buffer
}

// Write records response data before writing the response.
func (r *responseWriter) Write(body []byte) (int, error) {

	// record the response data
	if n, err := r.responseData.Write(body); err != nil {
		return n, err
	}

	// write the response
	return r.ResponseWriter.Write(body)

}
