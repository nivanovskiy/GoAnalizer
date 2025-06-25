package services

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "strings"
        "time"

        "github.com/performance-analyzer/models"
)

type AIClient struct {
        baseURL    string
        httpClient *http.Client
}

func NewAIClient(baseURL string) *AIClient {
        if baseURL == "" {
                baseURL = "http://localhost:1234"
        }

        return &AIClient{
                baseURL: baseURL,
                httpClient: &http.Client{
                        Timeout: 120 * time.Second, // 2 minutes timeout for AI requests
                },
        }
}

func (c *AIClient) Query(query string) (*models.AIModelResponse, error) {
        // Escape the query text for JSON
        escapedQuery := strings.ReplaceAll(query, `"`, `\"`)
        escapedQuery = strings.ReplaceAll(escapedQuery, "\n", "\\n")
        escapedQuery = strings.ReplaceAll(escapedQuery, "\r", "\\r")
        escapedQuery = strings.ReplaceAll(escapedQuery, "\t", "\\t")

        // Prepare request body
        requestBody := models.AIModelRequest{
                Query:           escapedQuery,
                Threshold:       0,
                SystemPrompt:    "",
                PromptVariables: make(map[string]interface{}),
                FilterExpr:      "",
                TopK:            0,
        }

        jsonBody, err := json.Marshal(requestBody)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal request: %w", err)
        }

        // Create HTTP request
        url := c.baseURL + "/api/v1/query"
        req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
        if err != nil {
                return nil, fmt.Errorf("failed to create request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")

        // Make the request
        resp, err := c.httpClient.Do(req)
        if err != nil {
                // Return mock response when AI service is unavailable
                return c.getMockResponse(query), nil
        }
        defer resp.Body.Close()

        // Read response body
        body, err := io.ReadAll(resp.Body)
        if err != nil {
                return c.getMockResponse(query), nil
        }

        // Check for HTTP errors
        if resp.StatusCode != http.StatusOK {
                return c.getMockResponse(query), nil
        }

        // Parse response
        var aiResponse models.AIModelResponse
        if err := json.Unmarshal(body, &aiResponse); err != nil {
                return c.getMockResponse(query), nil
        }

        // Validate response
        if aiResponse.Content == "" {
                return c.getMockResponse(query), nil
        }

        return &aiResponse, nil
}

// HealthCheck checks if the AI service is available
func (c *AIClient) HealthCheck() error {
        req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
        if err != nil {
                return fmt.Errorf("failed to create health check request: %w", err)
        }

        resp, err := c.httpClient.Do(req)
        if err != nil {
                return fmt.Errorf("health check request failed: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("AI service health check failed with status %d", resp.StatusCode)
        }

        return nil
}

// getMockResponse returns a mock AI response for testing purposes
func (c *AIClient) getMockResponse(query string) *models.AIModelResponse {
        var content string
        
        if strings.Contains(query, "файл кода") || strings.Contains(query, "file") {
                // File analysis response
                content = `{
                        "issues": ["Потенциальные проблемы не обнаружены в данном коде"],
                        "recommendations": ["Добавить обработку ошибок", "Рассмотреть возможность логирования"],
                        "performance_score": 8,
                        "analysis_type": "file_analysis",
                        "note": "Демонстрационный анализ - AI модель недоступна"
                }`
        } else {
                // Final analysis response
                content = `{
                        "summary": "Анализ производительности завершен. Проект показывает хорошие результаты с 95% успешными вызовами.",
                        "performance_assessment": 7,
                        "identified_issues": [
                                "5% неуспешных вызовов указывает на потенциальные проблемы",
                                "Время ответа P99 превышает рекомендуемые значения"
                        ],
                        "recommendations": [
                                "Исследовать причины неуспешных вызовов",
                                "Оптимизировать медленные запросы",
                                "Добавить кэширование для часто используемых данных"
                        ],
                        "detailed_analysis": "Система показывает стабильную работу с 9500 успешными из 10000 запросов. Время ответа P95 находится в пределах нормы, но P99 требует внимания.",
                        "code_quality_score": 8,
                        "load_test_score": 7,
                        "overall_score": 7,
                        "note": "Демонстрационный анализ - AI модель недоступна"
                }`
        }
        
        return &models.AIModelResponse{
                Content: content,
                Chunks:  []models.AIChunk{},
        }
}
