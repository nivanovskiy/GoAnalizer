package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriter wrapper to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// HTTPLogger middleware logs all HTTP requests and responses
func HTTPLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[HTTP] %v | %3d | %13v | %15s | %-7s %#v\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	})
}

// DetailedHTTPLogger logs detailed request/response information
func DetailedHTTPLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Log request
		logRequest(c)
		
		// Wrap response writer to capture response body
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Log response
		latency := time.Since(start)
		logResponse(c, blw, latency)
	}
}

func logRequest(c *gin.Context) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Format request body for logging
	var requestBody interface{}
	if len(bodyBytes) > 0 && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			requestBody = string(bodyBytes)
		}
	}

	log.Printf("=== INCOMING HTTP REQUEST ===")
	log.Printf("Method: %s", c.Request.Method)
	log.Printf("URL: %s", c.Request.URL.String())
	log.Printf("Headers: %v", formatHeaders(c.Request.Header))
	
	if requestBody != nil {
		requestBodyJSON, _ := json.MarshalIndent(requestBody, "", "  ")
		log.Printf("Body: %s", string(requestBodyJSON))
	}
	log.Printf("Client IP: %s", c.ClientIP())
	log.Printf("User Agent: %s", c.Request.UserAgent())
}

func logResponse(c *gin.Context, blw *responseWriter, latency time.Duration) {
	log.Printf("=== OUTGOING HTTP RESPONSE ===")
	log.Printf("Status: %d", c.Writer.Status())
	log.Printf("Latency: %v", latency)
	log.Printf("Response Headers: %v", formatHeaders(c.Writer.Header()))
	
	// Log response body if it's JSON
	responseBody := blw.body.String()
	if responseBody != "" {
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(responseBody), &jsonBody); err == nil {
			responseBodyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
			log.Printf("Response Body: %s", string(responseBodyJSON))
		} else {
			log.Printf("Response Body: %s", responseBody)
		}
	}
	log.Printf("===============================")
}

func formatHeaders(headers map[string][]string) map[string]string {
	formatted := make(map[string]string)
	for key, values := range headers {
		// Hide sensitive headers
		if strings.ToLower(key) == "authorization" || strings.ToLower(key) == "cookie" {
			formatted[key] = "[HIDDEN]"
		} else {
			formatted[key] = strings.Join(values, ", ")
		}
	}
	return formatted
}