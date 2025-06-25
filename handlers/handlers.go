package handlers

import (
        "context"
        "encoding/json"
        "net/http"
        "time"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/performance-analyzer/models"
        "github.com/performance-analyzer/services"
        "github.com/performance-analyzer/utils"
)

type Handler struct {
        db       *pgxpool.Pool
        analyzer *services.Analyzer
}

func New(db *pgxpool.Pool, analyzer *services.Analyzer) *Handler {
        return &Handler{
                db:       db,
                analyzer: analyzer,
        }
}

func (h *Handler) InitAnalyze(c *gin.Context) {
        tenant := c.Param("tenant")
        repo := c.Param("repo")
        uuidParam := c.Param("uuid")

        // Validate UUID
        projectUUID, err := uuid.Parse(uuidParam)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
                return
        }

        // Validate tenant and repo
        if err := utils.ValidateString(tenant, 1, 255); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant: " + err.Error()})
                return
        }
        if err := utils.ValidateString(repo, 1, 255); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repo: " + err.Error()})
                return
        }

        // Parse request body
        var req models.InitAnalyzeRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
                return
        }

        // Validate request fields
        if err := utils.ValidateString(req.Language, 1, 100); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid language: " + err.Error()})
                return
        }
        if err := utils.ValidateString(req.TestingTool, 1, 100); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid testing tool: " + err.Error()})
                return
        }

        // Check if project already exists
        var existingID int
        err = h.db.QueryRow(context.Background(), 
                "SELECT id FROM projects WHERE uuid = $1", projectUUID).Scan(&existingID)
        if err == nil {
                c.JSON(http.StatusConflict, gin.H{"error": "Project with this UUID already exists"})
                return
        }

        // Insert project into database
        query := `
                INSERT INTO projects (tenant, repo, uuid, language, testing_tool, project_info, status)
                VALUES ($1, $2, $3, $4, $5, $6, 'initialized')
                RETURNING id`
        
        var projectID int
        err = h.db.QueryRow(context.Background(), query, 
                tenant, repo, projectUUID, req.Language, req.TestingTool, req.ProjectInfo).Scan(&projectID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project: " + err.Error()})
                return
        }

        // Initialize analysis result record
        _, err = h.db.Exec(context.Background(),
                "INSERT INTO analysis_results (project_uuid, status) VALUES ($1, 'pending')",
                projectUUID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize analysis: " + err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{
                "message":    "Analysis initialized successfully",
                "project_id": projectID,
                "uuid":       projectUUID,
        })
}

func (h *Handler) SendFile(c *gin.Context) {
        uuidParam := c.Param("uuid")

        // Validate UUID
        projectUUID, err := uuid.Parse(uuidParam)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
                return
        }

        // Parse request body
        var req models.SendFileRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
                return
        }

        // Validate filename and content
        if err := utils.ValidateString(req.Filename, 1, 500); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename: " + err.Error()})
                return
        }
        if len(req.Content) == 0 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "File content cannot be empty"})
                return
        }

        // Check if project exists
        var projectExists bool
        err = h.db.QueryRow(context.Background(),
                "SELECT EXISTS(SELECT 1 FROM projects WHERE uuid = $1)", projectUUID).Scan(&projectExists)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
                return
        }
        if !projectExists {
                c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
                return
        }

        // Request AI analysis for the file
        fileAnalysis, err := h.analyzer.AnalyzeFile(req.Content)
        if err != nil {
                // Log error but continue - we'll store the file without analysis
                errorAnalysis := map[string]interface{}{
                        "error":       "AI analysis failed",
                        "message":     err.Error(),
                        "analyzed_at": time.Now(),
                        "file_size":   len(req.Content),
                }
                fileAnalysisBytes, _ := json.Marshal(errorAnalysis)
                fileAnalysis = json.RawMessage(fileAnalysisBytes)
        }

        // Insert or update file
        query := `
                INSERT INTO project_files (project_uuid, filename, content, file_analysis)
                VALUES ($1, $2, $3, $4)
                ON CONFLICT (project_uuid, filename)
                DO UPDATE SET content = EXCLUDED.content, file_analysis = EXCLUDED.file_analysis`
        
        _, err = h.db.Exec(context.Background(), query,
                projectUUID, req.Filename, req.Content, fileAnalysis)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "message":  "File received and analyzed successfully",
                "filename": req.Filename,
                "uuid":     projectUUID,
        })
}

func (h *Handler) SendResults(c *gin.Context) {
        uuidParam := c.Param("uuid")

        // Validate UUID
        projectUUID, err := uuid.Parse(uuidParam)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
                return
        }

        // Parse request body
        var req models.SendResultsRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
                return
        }

        // Validate required fields
        if req.SuccessfulCalls < 0 || req.FailedCalls < 0 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Call counts cannot be negative"})
                return
        }

        // Check if project exists
        var projectExists bool
        err = h.db.QueryRow(context.Background(),
                "SELECT EXISTS(SELECT 1 FROM projects WHERE uuid = $1)", projectUUID).Scan(&projectExists)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
                return
        }
        if !projectExists {
                c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
                return
        }

        // Prepare raw results JSON
        rawResults := map[string]interface{}{
                "response_time_p95":           req.ResponseTimeP95,
                "response_time_p99":           req.ResponseTimeP99,
                "successful_calls":            req.SuccessfulCalls,
                "failed_calls":                req.FailedCalls,
                "nonfunctional_requirements": req.NonfunctionalRequirements,
                "raw_results":                req.RawResults,
                "received_at":                time.Now(),
        }
        rawResultsJSON, _ := json.Marshal(rawResults)

        // Insert test results
        query := `
                INSERT INTO test_results (project_uuid, response_time_p95, response_time_p99, 
                                         successful_calls, failed_calls, nonfunctional_requirements, raw_results)
                VALUES ($1, $2, $3, $4, $5, $6, $7)`
        
        _, err = h.db.Exec(context.Background(), query,
                projectUUID, req.ResponseTimeP95, req.ResponseTimeP99,
                req.SuccessfulCalls, req.FailedCalls, req.NonfunctionalRequirements, rawResultsJSON)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save test results: " + err.Error()})
                return
        }

        // Update project status to indicate results received
        _, err = h.db.Exec(context.Background(),
                "UPDATE projects SET status = 'results_received', updated_at = CURRENT_TIMESTAMP WHERE uuid = $1",
                projectUUID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project status: " + err.Error()})
                return
        }

        // Trigger final analysis
        h.analyzer.TriggerFinalAnalysis(projectUUID)

        c.JSON(http.StatusOK, gin.H{
                "message": "Test results received successfully",
                "uuid":    projectUUID,
        })
}

func (h *Handler) GetAnalyzeResults(c *gin.Context) {
        uuidParam := c.Param("uuid")

        // Validate UUID
        projectUUID, err := uuid.Parse(uuidParam)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
                return
        }

        // Get analysis result
        var result models.AnalysisResult
        query := `
                SELECT id, project_uuid, final_analysis, status, error_message, created_at, completed_at
                FROM analysis_results
                WHERE project_uuid = $1
                ORDER BY created_at DESC
                LIMIT 1`
        
        err = h.db.QueryRow(context.Background(), query, projectUUID).Scan(
                &result.ID, &result.ProjectUUID, &result.FinalAnalysis,
                &result.Status, &result.ErrorMessage, &result.CreatedAt, &result.CompletedAt)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
                return
        }

        // Check status
        switch result.Status {
        case "pending", "processing":
                c.JSON(http.StatusAccepted, gin.H{
                        "status":  result.Status,
                        "message": "Analysis is still in progress",
                        "uuid":    projectUUID,
                })
                return
        case "completed":
                // Return full analysis results
                var analysisData interface{}
                if result.FinalAnalysis != nil {
                        json.Unmarshal(result.FinalAnalysis, &analysisData)
                }

                response := gin.H{
                        "uuid":         projectUUID,
                        "status":       result.Status,
                        "analysis":     analysisData,
                        "completed_at": result.CompletedAt,
                }

                c.JSON(http.StatusOK, response)
                return
        case "failed":
                errorMsg := "Analysis failed"
                if result.ErrorMessage != nil {
                        errorMsg = *result.ErrorMessage
                }
                c.JSON(http.StatusInternalServerError, gin.H{
                        "status":  result.Status,
                        "error":   errorMsg,
                        "uuid":    projectUUID,
                })
                return
        default:
                c.JSON(http.StatusInternalServerError, gin.H{
                        "error": "Unknown analysis status",
                        "uuid":  projectUUID,
                })
                return
        }
}
