package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggedHTTPClient wraps http.Client with logging
type LoggedHTTPClient struct {
	client *http.Client
}

// NewLoggedHTTPClient creates a new HTTP client with logging
func NewLoggedHTTPClient() *LoggedHTTPClient {
	return &LoggedHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Post makes a POST request with detailed logging
func (c *LoggedHTTPClient) Post(url string, body interface{}) (*http.Response, error) {
	return c.doRequest("POST", url, body)
}

// Get makes a GET request with detailed logging
func (c *LoggedHTTPClient) Get(url string) (*http.Response, error) {
	return c.doRequest("GET", url, nil)
}

// Put makes a PUT request with detailed logging
func (c *LoggedHTTPClient) Put(url string, body interface{}) (*http.Response, error) {
	return c.doRequest("PUT", url, body)
}

// Delete makes a DELETE request with detailed logging
func (c *LoggedHTTPClient) Delete(url string) (*http.Response, error) {
	return c.doRequest("DELETE", url, nil)
}

func (c *LoggedHTTPClient) doRequest(method, url string, body interface{}) (*http.Response, error) {
	start := time.Now()
	
	// Prepare request body
	var requestBody io.Reader
	var bodyBytes []byte
	
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			log.Printf("Failed to marshal request body: %v", err)
			return nil, err
		}
		requestBody = bytes.NewBuffer(bodyBytes)
	}
	
	// Create request
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return nil, err
	}
	
	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "Performance-Analyzer/1.0")
	
	// Log outgoing request
	c.logOutgoingRequest(req, bodyBytes)
	
	// Make request
	resp, err := c.client.Do(req)
	latency := time.Since(start)
	
	if err != nil {
		log.Printf("HTTP request failed: %v (latency: %v)", err, latency)
		return nil, err
	}
	
	// Log response
	c.logIncomingResponse(resp, latency)
	
	return resp, nil
}

func (c *LoggedHTTPClient) logOutgoingRequest(req *http.Request, bodyBytes []byte) {
	log.Printf("=== OUTGOING HTTP REQUEST ===")
	log.Printf("Method: %s", req.Method)
	log.Printf("URL: %s", req.URL.String())
	log.Printf("Headers: %v", formatRequestHeaders(req.Header))
	
	if len(bodyBytes) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
			requestBodyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
			log.Printf("Request Body: %s", string(requestBodyJSON))
		} else {
			log.Printf("Request Body: %s", string(bodyBytes))
		}
	}
}

func (c *LoggedHTTPClient) logIncomingResponse(resp *http.Response, latency time.Duration) {
	log.Printf("=== INCOMING HTTP RESPONSE ===")
	log.Printf("Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("Latency: %v", latency)
	log.Printf("Response Headers: %v", formatResponseHeaders(resp.Header))
	
	// Read and log response body
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body: %v", err)
		} else {
			// Create a new reader for the response body
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			if len(bodyBytes) > 0 {
				var jsonBody interface{}
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					responseBodyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
					log.Printf("Response Body: %s", string(responseBodyJSON))
				} else {
					log.Printf("Response Body: %s", string(bodyBytes))
				}
			}
		}
	}
	log.Printf("=============================")
}

func formatRequestHeaders(headers http.Header) map[string]string {
	formatted := make(map[string]string)
	for key, values := range headers {
		// Hide sensitive headers
		if key == "Authorization" || key == "Cookie" || key == "X-Api-Key" {
			formatted[key] = "[HIDDEN]"
		} else {
			formatted[key] = fmt.Sprintf("%v", values)
		}
	}
	return formatted
}

func formatResponseHeaders(headers http.Header) map[string]string {
	formatted := make(map[string]string)
	for key, values := range headers {
		formatted[key] = fmt.Sprintf("%v", values)
	}
	return formatted
}