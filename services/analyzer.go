package services

import (
        "context"
        "encoding/json"
        "fmt"
        "log"
        "strings"
        "time"

        "github.com/google/uuid"
        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/performance-analyzer/models"
)

type Analyzer struct {
        db       *pgxpool.Pool
        aiClient *AIClient
        queue    chan uuid.UUID
}

func NewAnalyzer(db *pgxpool.Pool, aiClient *AIClient) *Analyzer {
        return &Analyzer{
                db:       db,
                aiClient: aiClient,
                queue:    make(chan uuid.UUID, 100), // Buffer for 100 analysis requests
        }
}

func (a *Analyzer) StartBackgroundProcessor() {
        log.Println("Starting background analyzer processor...")
        for projectUUID := range a.queue {
                log.Printf("Processing analysis for project %s", projectUUID)
                a.processAnalysis(projectUUID)
        }
}

func (a *Analyzer) TriggerFinalAnalysis(projectUUID uuid.UUID) {
        select {
        case a.queue <- projectUUID:
                log.Printf("Queued analysis for project %s", projectUUID)
        default:
                log.Printf("Analysis queue is full, skipping project %s", projectUUID)
        }
}

func (a *Analyzer) AnalyzeFile(content string) (json.RawMessage, error) {
        prompt := fmt.Sprintf(`Проанализируйте следующий файл кода как эксперт по тестированию производительности. 
Укажите потенциальные проблемы производительности, узкие места, и рекомендации по оптимизации.
Ответьте в формате JSON с полями: issues (проблемы), recommendations (рекомендации), performance_score (оценка от 1 до 10).

Код файла:
%s`, content)

        response, err := a.aiClient.Query(prompt)
        if err != nil {
                return nil, fmt.Errorf("AI analysis failed: %w", err)
        }

        // Parse and structure the response
        analysisResult := map[string]interface{}{
                "ai_response":  response.Content,
                "analyzed_at":  time.Now(),
                "file_size":    len(content),
                "analysis_type": "file_analysis",
        }

        resultJSON, err := json.Marshal(analysisResult)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal analysis result: %w", err)
        }

        return json.RawMessage(resultJSON), nil
}

func (a *Analyzer) processAnalysis(projectUUID uuid.UUID) {
        // Update status to processing
        _, err := a.db.Exec(context.Background(),
                "UPDATE analysis_results SET status = 'processing' WHERE project_uuid = $1",
                projectUUID)
        if err != nil {
                log.Printf("Failed to update analysis status for %s: %v", projectUUID, err)
                return
        }

        // Get project information
        project, err := a.getProject(projectUUID)
        if err != nil {
                a.markAnalysisFailed(projectUUID, fmt.Sprintf("Failed to get project: %v", err))
                return
        }

        // Get project files
        files, err := a.getProjectFiles(projectUUID)
        if err != nil {
                a.markAnalysisFailed(projectUUID, fmt.Sprintf("Failed to get project files: %v", err))
                return
        }

        // Get test results
        testResults, err := a.getTestResults(projectUUID)
        if err != nil {
                a.markAnalysisFailed(projectUUID, fmt.Sprintf("Failed to get test results: %v", err))
                return
        }

        // Perform comprehensive analysis
        finalAnalysis, err := a.performFinalAnalysis(project, files, testResults)
        if err != nil {
                a.markAnalysisFailed(projectUUID, fmt.Sprintf("AI analysis failed: %v", err))
                return
        }

        // Save analysis results
        now := time.Now()
        _, err = a.db.Exec(context.Background(),
                `UPDATE analysis_results 
                 SET final_analysis = $1, status = 'completed', completed_at = $2 
                 WHERE project_uuid = $3`,
                finalAnalysis, now, projectUUID)
        if err != nil {
                log.Printf("Failed to save analysis results for %s: %v", projectUUID, err)
                a.markAnalysisFailed(projectUUID, fmt.Sprintf("Failed to save results: %v", err))
                return
        }

        log.Printf("Successfully completed analysis for project %s", projectUUID)
}

func (a *Analyzer) getProject(projectUUID uuid.UUID) (*models.Project, error) {
        var project models.Project
        query := `
                SELECT id, tenant, repo, uuid, language, testing_tool, project_info, 
                       files_count, received_files_count, has_test_results, status, created_at, updated_at
                FROM projects WHERE uuid = $1`
        
        err := a.db.QueryRow(context.Background(), query, projectUUID).Scan(
                &project.ID, &project.Tenant, &project.Repo, &project.UUID,
                &project.Language, &project.TestingTool, &project.ProjectInfo,
                &project.FilesCount, &project.ReceivedFilesCount, &project.HasTestResults,
                &project.Status, &project.CreatedAt, &project.UpdatedAt)
        
        return &project, err
}

func (a *Analyzer) getProjectFiles(projectUUID uuid.UUID) ([]models.ProjectFile, error) {
        query := `
                SELECT id, project_uuid, filename, content, file_analysis, created_at
                FROM project_files WHERE project_uuid = $1`
        
        rows, err := a.db.Query(context.Background(), query, projectUUID)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var files []models.ProjectFile
        for rows.Next() {
                var file models.ProjectFile
                err := rows.Scan(&file.ID, &file.ProjectUUID, &file.Filename,
                        &file.Content, &file.FileAnalysis, &file.CreatedAt)
                if err != nil {
                        return nil, err
                }
                files = append(files, file)
        }

        return files, rows.Err()
}

func (a *Analyzer) getTestResults(projectUUID uuid.UUID) (*models.TestResults, error) {
        var testResult models.TestResults
        query := `
                SELECT id, project_uuid, response_time_p95, response_time_p99, 
                       successful_calls, failed_calls, nonfunctional_requirements, raw_results, created_at
                FROM test_results WHERE project_uuid = $1 ORDER BY created_at DESC LIMIT 1`
        
        err := a.db.QueryRow(context.Background(), query, projectUUID).Scan(
                &testResult.ID, &testResult.ProjectUUID, &testResult.ResponseTimeP95,
                &testResult.ResponseTimeP99, &testResult.SuccessfulCalls, &testResult.FailedCalls,
                &testResult.NonfunctionalRequirements, &testResult.RawResults, &testResult.CreatedAt)
        
        return &testResult, err
}

func (a *Analyzer) performFinalAnalysis(project *models.Project, files []models.ProjectFile, testResults *models.TestResults) (json.RawMessage, error) {
        // Prepare comprehensive analysis prompt
        var filesSummary strings.Builder
        filesSummary.WriteString("Файлы проекта:\n")
        for _, file := range files {
                filesSummary.WriteString(fmt.Sprintf("- %s (размер: %d символов)\n", file.Filename, len(file.Content)))
                if file.FileAnalysis != nil {
                        filesSummary.WriteString(fmt.Sprintf("  Анализ: %s\n", string(file.FileAnalysis)))
                }
        }

        testSummary := fmt.Sprintf(`
Результаты тестирования:
- Успешные вызовы: %d
- Неуспешные вызовы: %d
- 95-й перцентиль времени ответа: %s
- 99-й перцентиль времени ответа: %s
- Нефункциональные требования: %s
- Дополнительные результаты: %s`,
                testResults.SuccessfulCalls, testResults.FailedCalls,
                string(testResults.ResponseTimeP95), string(testResults.ResponseTimeP99),
                string(testResults.NonfunctionalRequirements), string(testResults.RawResults))

        prompt := fmt.Sprintf(`Проанализируйте результаты тестирования производительности как эксперт.
Объясните простым языком пользователю:

Информация о проекте:
- Язык: %s
- Инструмент тестирования: %s
- Дополнительная информация: %s

%s

%s

Предоставьте анализ в формате JSON со следующими полями:
- summary: краткое резюме на русском языке
- performance_assessment: общая оценка производительности (1-10)
- identified_issues: список выявленных проблем
- recommendations: список рекомендаций по улучшению
- detailed_analysis: подробный анализ результатов
- code_quality_score: оценка качества кода (1-10)
- load_test_score: оценка результатов нагрузочного тестирования (1-10)
- overall_score: общая оценка проекта (1-10)

Используйте простой язык для объяснения технических вопросов.`,
                project.Language, project.TestingTool, string(project.ProjectInfo),
                filesSummary.String(), testSummary)

        response, err := a.aiClient.Query(prompt)
        if err != nil {
                return nil, err
        }

        // Structure the final analysis
        finalAnalysis := map[string]interface{}{
                "ai_analysis":    response.Content,
                "project_info": map[string]interface{}{
                        "tenant":       project.Tenant,
                        "repo":         project.Repo,
                        "language":     project.Language,
                        "testing_tool": project.TestingTool,
                },
                "files_count":    len(files),
                "test_summary": map[string]interface{}{
                        "successful_calls": testResults.SuccessfulCalls,
                        "failed_calls":     testResults.FailedCalls,
                        "total_calls":      testResults.SuccessfulCalls + testResults.FailedCalls,
                },
                "analysis_metadata": map[string]interface{}{
                        "analyzed_at":      time.Now(),
                        "analysis_version": "1.0",
                },
        }

        finalAnalysisJSON, err := json.Marshal(finalAnalysis)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal final analysis: %w", err)
        }

        return json.RawMessage(finalAnalysisJSON), nil
}

func (a *Analyzer) markAnalysisFailed(projectUUID uuid.UUID, errorMsg string) {
        _, err := a.db.Exec(context.Background(),
                `UPDATE analysis_results 
                 SET status = 'failed', error_message = $1 
                 WHERE project_uuid = $2`,
                errorMsg, projectUUID)
        if err != nil {
                log.Printf("Failed to mark analysis as failed for %s: %v", projectUUID, err)
        }
}
