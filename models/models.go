package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          int             `json:"id" db:"id"`
	Tenant      string          `json:"tenant" db:"tenant"`
	Repo        string          `json:"repo" db:"repo"`
	UUID        uuid.UUID       `json:"uuid" db:"uuid"`
	Language    string          `json:"language" db:"language"`
	TestingTool string          `json:"testing_tool" db:"testing_tool"`
	ProjectInfo json.RawMessage `json:"project_info" db:"project_info"`
	Status      string          `json:"status" db:"status"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type ProjectFile struct {
	ID           int             `json:"id" db:"id"`
	ProjectUUID  uuid.UUID       `json:"project_uuid" db:"project_uuid"`
	Filename     string          `json:"filename" db:"filename"`
	Content      string          `json:"content" db:"content"`
	FileAnalysis json.RawMessage `json:"file_analysis" db:"file_analysis"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type TestResults struct {
	ID                        int             `json:"id" db:"id"`
	ProjectUUID               uuid.UUID       `json:"project_uuid" db:"project_uuid"`
	ResponseTimeP95           json.RawMessage `json:"response_time_p95" db:"response_time_p95"`
	ResponseTimeP99           json.RawMessage `json:"response_time_p99" db:"response_time_p99"`
	SuccessfulCalls           int             `json:"successful_calls" db:"successful_calls"`
	FailedCalls               int             `json:"failed_calls" db:"failed_calls"`
	NonfunctionalRequirements json.RawMessage `json:"nonfunctional_requirements" db:"nonfunctional_requirements"`
	RawResults                json.RawMessage `json:"raw_results" db:"raw_results"`
	CreatedAt                 time.Time       `json:"created_at" db:"created_at"`
}

type AnalysisResult struct {
	ID            int             `json:"id" db:"id"`
	ProjectUUID   uuid.UUID       `json:"project_uuid" db:"project_uuid"`
	FinalAnalysis json.RawMessage `json:"final_analysis" db:"final_analysis"`
	Status        string          `json:"status" db:"status"`
	ErrorMessage  *string         `json:"error_message" db:"error_message"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	CompletedAt   *time.Time      `json:"completed_at" db:"completed_at"`
}

// Request/Response models
type InitAnalyzeRequest struct {
	Language    string          `json:"language"`
	TestingTool string          `json:"testing_tool"`
	ProjectInfo json.RawMessage `json:"project_info"`
}

type SendFileRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type SendResultsRequest struct {
	ResponseTimeP95           json.RawMessage `json:"response_time_p95"`
	ResponseTimeP99           json.RawMessage `json:"response_time_p99"`
	SuccessfulCalls           int             `json:"successful_calls"`
	FailedCalls               int             `json:"failed_calls"`
	NonfunctionalRequirements json.RawMessage `json:"nonfunctional_requirements"`
	RawResults                json.RawMessage `json:"raw_results"`
}

// AI Model API structures
type AIModelRequest struct {
	Query           string                 `json:"query"`
	Threshold       int                    `json:"threshold"`
	SystemPrompt    string                 `json:"system_prompt"`
	PromptVariables map[string]interface{} `json:"prompt_variables"`
	FilterExpr      string                 `json:"filter_expr"`
	TopK            int                    `json:"top_k"`
}

type AIModelResponse struct {
	Content string     `json:"content"`
	Chunks  []AIChunk `json:"chunks"`
}

type AIChunk struct {
	Metadata   AIMetadata `json:"metadata"`
	Content    string     `json:"content"`
	Embeddings []float64  `json:"embeddings"`
	Score      float64    `json:"score"`
}

type AIMetadata struct {
	DocID        string                 `json:"doc_id"`
	Hash         string                 `json:"hash"`
	DatasourceID int                    `json:"datasource_id"`
	PipelineID   int                    `json:"pipeline_id"`
	ChunkIndex   int                    `json:"chunk_index"`
	ProjectID    int                    `json:"project_id"`
	Additional   map[string]interface{} `json:"additionalProp1"`
}
