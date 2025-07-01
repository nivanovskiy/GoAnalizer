# Performance Analyzer API - Примеры запросов

## Базовый URL
```
http://localhost:5000
```

## 1. Проверка состояния сервиса

### GET /health
```bash
curl -X GET http://localhost:5000/health
```

**Ответ:**
```json
{
  "status": "healthy"
}
```

### GET / (Документация API)
```bash
curl -X GET http://localhost:5000/
```

**Ответ:**
```json
{
  "service": "Performance Analyzer API",
  "version": "1.0.0",
  "status": "running",
  "endpoints": {
    "POST /initAnalize/{tenant}/{repo}/{uuid}": "Initialize analysis pipeline",
    "POST /sendFile/{uuid}": "Upload project file for analysis",
    "POST /sendResults/{uuid}": "Submit performance test results",
    "GET /getAnalizeResults/{uuid}": "Get analysis results",
    "GET /health": "Health check"
  },
  "description": "REST API for performance testing analysis with AI-powered insights"
}
```

## 2. Инициализация анализа

### POST /initAnalize/{tenant}/{repo}/{uuid}

```bash
curl -X POST http://localhost:5000/initAnalize/my-company/web-app/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "language": "Go",
    "testing_tool": "k6",
    "files_count": 3,
    "project_info": {
      "description": "Веб-приложение для онлайн магазина",
      "version": "2.1.0",
      "team": "Backend Team",
      "expected_load": "1000 RPS"
    }
  }'
```

**Ответ:**
```json
{
  "message": "Analysis initialized successfully",
  "project_id": 1,
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "files_count": 3
}
```

## 3. Загрузка файлов проекта

### POST /sendFile/{uuid}

#### Пример 1: Главный сервер
```bash
curl -X POST http://localhost:5000/sendFile/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "main.go",
    "content": "package main\n\nimport (\n    \"log\"\n    \"net/http\"\n    \"github.com/gin-gonic/gin\"\n    \"github.com/your-org/shop/handlers\"\n    \"github.com/your-org/shop/database\"\n)\n\nfunc main() {\n    db := database.Connect()\n    defer db.Close()\n    \n    r := gin.Default()\n    \n    // API routes\n    api := r.Group(\"/api/v1\")\n    {\n        api.GET(\"/products\", handlers.GetProducts(db))\n        api.POST(\"/orders\", handlers.CreateOrder(db))\n        api.GET(\"/users/:id\", handlers.GetUser(db))\n    }\n    \n    log.Println(\"Starting server on :8080\")\n    r.Run(\":8080\")\n}"
  }'
```

#### Пример 2: Обработчик продуктов
```bash
curl -X POST http://localhost:5000/sendFile/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "handlers/products.go",
    "content": "package handlers\n\nimport (\n    \"net/http\"\n    \"strconv\"\n    \"github.com/gin-gonic/gin\"\n    \"github.com/your-org/shop/models\"\n    \"gorm.io/gorm\"\n)\n\nfunc GetProducts(db *gorm.DB) gin.HandlerFunc {\n    return func(c *gin.Context) {\n        // Потенциальная проблема: нет пагинации\n        var products []models.Product\n        \n        // Медленный запрос без индексов\n        result := db.Where(\"status = ? AND category LIKE ?\", \"active\", \"%\"+c.Query(\"search\")+\"%\").Find(&products)\n        \n        if result.Error != nil {\n            c.JSON(http.StatusInternalServerError, gin.H{\"error\": \"Database error\"})\n            return\n        }\n        \n        // Возвращаем все продукты без лимита\n        c.JSON(http.StatusOK, gin.H{\"products\": products, \"count\": len(products)})\n    }\n}"
  }'
```

#### Пример 3: База данных
```bash
curl -X POST http://localhost:5000/sendFile/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "database/connection.go",
    "content": "package database\n\nimport (\n    \"log\"\n    \"gorm.io/driver/postgres\"\n    \"gorm.io/gorm\"\n)\n\nfunc Connect() *gorm.DB {\n    // Проблема: нет пула соединений\n    dsn := \"host=localhost user=shop password=123456 dbname=shop_db port=5432 sslmode=disable\"\n    \n    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})\n    if err != nil {\n        log.Fatal(\"Failed to connect to database:\", err)\n    }\n    \n    // Проблема: нет настройки connection pool\n    return db\n}"
  }'
```

**Ответ для каждого файла:**
```json
{
  "message": "File received and analyzed successfully",
  "filename": "main.go",
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "received_files_count": 1,
  "total_files_count": 3,
  "ready_for_analysis": false
}
```

**Ответ для последнего файла (если все файлы получены и есть результаты тестов):**
```json
{
  "message": "File received and analyzed successfully",
  "filename": "database/connection.go",
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "received_files_count": 3,
  "total_files_count": 3,
  "ready_for_analysis": true
}
```

## 4. Отправка результатов тестирования

### POST /sendResults/{uuid}

#### Пример реалистичных результатов нагрузочного тестирования
```bash
curl -X POST http://localhost:5000/sendResults/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "response_time_p95": {
      "GET_api_v1_products": 1200,
      "POST_api_v1_orders": 2500,
      "GET_api_v1_users": 800
    },
    "response_time_p99": {
      "GET_api_v1_products": 2100,
      "POST_api_v1_orders": 4500,
      "GET_api_v1_users": 1500
    },
    "successful_calls": 7800,
    "failed_calls": 2200,
    "nonfunctional_requirements": {
      "max_response_time_ms": 500,
      "target_throughput_rps": 1000,
      "target_error_rate_percent": 1,
      "actual_error_rate_percent": 22
    },
    "raw_results": {
      "test_duration": "10m",
      "total_requests": 10000,
      "requests_per_second": 16.67,
      "data_transferred": "15.2 MB",
      "errors": [
        {
          "type": "timeout",
          "count": 1500,
          "percentage": 15
        },
        {
          "type": "connection_refused",
          "count": 400,
          "percentage": 4
        },
        {
          "type": "500_internal_server_error",
          "count": 300,
          "percentage": 3
        }
      ],
      "response_time_distribution": {
        "min": 45,
        "max": 8500,
        "avg": 1250,
        "p50": 980,
        "p90": 1800,
        "p95": 2200,
        "p99": 4200
      }
    }
  }'
```

**Ответ (если не все файлы получены):**
```json
{
  "message": "Test results received successfully",
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "received_files_count": 2,
  "total_files_count": 3,
  "ready_for_analysis": false
}
```

**Ответ (если все файлы получены - запускается анализ):**
```json
{
  "message": "Test results received successfully",
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "received_files_count": 3,
  "total_files_count": 3,
  "ready_for_analysis": true
}
```

## 5. Получение результатов анализа

### GET /getAnalizeResults/{uuid}

```bash
curl -X GET http://localhost:5000/getAnalizeResults/123e4567-e89b-12d3-a456-426614174000
```

#### Возможные ответы:

**Анализ в процессе (код 202):**
```json
{
  "status": "processing",
  "message": "Analysis is still in progress",
  "uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Анализ завершен (код 200):**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "status": "completed",
  "completed_at": "2025-06-25T10:46:33.320436Z",
  "analysis": {
    "ai_analysis": {
      "summary": "Анализ производительности завершен. Проект показывает серьезные проблемы с 22% неуспешными вызовами.",
      "performance_assessment": 4,
      "identified_issues": [
        "22% неуспешных вызовов значительно превышает допустимый уровень",
        "Время ответа P99 в 4+ раза превышает целевые значения",
        "Отсутствует пагинация в запросах продуктов",
        "Не настроен пул соединений с базой данных",
        "Медленные SQL запросы без индексов"
      ],
      "recommendations": [
        "Добавить пагинацию для всех списочных запросов",
        "Настроить connection pool для базы данных",
        "Добавить индексы для часто используемых полей поиска",
        "Реализовать кэширование для статических данных",
        "Добавить мониторинг и алерты на высокое время ответа",
        "Оптимизировать SQL запросы с EXPLAIN ANALYZE"
      ],
      "detailed_analysis": "Система испытывает серьезные проблемы производительности. 22% ошибок недопустимо для продакшна. Основные проблемы связаны с неэффективными запросами к базе данных и отсутствием базовых оптимизаций.",
      "code_quality_score": 5,
      "load_test_score": 3,
      "overall_score": 4
    },
    "project_info": {
      "tenant": "my-company",
      "repo": "web-app",
      "language": "Go",
      "testing_tool": "k6"
    },
    "files_count": 3,
    "test_summary": {
      "successful_calls": 7800,
      "failed_calls": 2200,
      "total_calls": 10000
    },
    "analysis_metadata": {
      "analyzed_at": "2025-06-25T10:46:33.320414852Z",
      "analysis_version": "1.0"
    }
  }
}
```

**Анализ с ошибкой (код 500):**
```json
{
  "status": "failed",
  "error": "AI analysis failed: connection timeout",
  "uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Полный пример workflow

```bash
# 1. Проверка сервиса
curl -X GET http://localhost:5000/health

# 2. Инициализация проекта (указываем количество файлов)
curl -X POST http://localhost:5000/initAnalize/test-company/my-api/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"language": "Go", "testing_tool": "k6", "files_count": 2, "project_info": {"version": "1.0"}}'

# 3. Загрузка файлов (по одному, должно соответствовать files_count)
# Файл 1
curl -X POST http://localhost:5000/sendFile/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"filename": "main.go", "content": "package main\n\nfunc main() {\n    // код сервера\n}"}'

# Файл 2 (последний - после него и sendResults запустится анализ)
curl -X POST http://localhost:5000/sendFile/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"filename": "handlers.go", "content": "package main\n\nfunc handler() {\n    // обработчики\n}"}'

# 4. Отправка результатов тестирования (анализ запустится если все файлы получены)
curl -X POST http://localhost:5000/sendResults/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"response_time_p95": {"api": 150}, "response_time_p99": {"api": 250}, "successful_calls": 9000, "failed_calls": 1000, "nonfunctional_requirements": {}, "raw_results": {}}'

# 5. Получение результатов (может потребоваться подождать)
curl -X GET http://localhost:5000/getAnalizeResults/550e8400-e29b-41d4-a716-446655440000
```

## Важные изменения в логике

1. **Обязательное указание количества файлов**: При инициализации обязательно указывается `files_count`
2. **Автоматический запуск анализа**: Анализ запускается только когда:
   - Получены все файлы (received_files_count >= files_count)
   - Получены результаты тестирования (has_test_results = true)
3. **Отслеживание прогресса**: Каждый ответ содержит:
   - `received_files_count` - количество полученных файлов
   - `total_files_count` - общее количество ожидаемых файлов
   - `ready_for_analysis` - готовность к запуску анализа
4. **Защита от дублирования**: Файлы с одинаковым именем перезаписываются, счетчик не увеличивается

## HTTP Логирование

Приложение теперь ведет детальное логирование всех HTTP запросов и ответов:

### Входящие запросы (API сервера)
- Метод HTTP, URL, заголовки
- Тело запроса в JSON формате
- IP адрес клиента и User Agent
- Время отклика и статус ответа
- Тело ответа в JSON формате

### Исходящие запросы (к AI сервису)
- Детальная информация о запросах к внешнему AI API
- Логирование запросов и ответов с временными метками
- Обработка ошибок с fallback на mock ответы

Все логи выводятся в консоль сервера в структурированном формате для удобного мониторинга и отладки.

## Инструменты для тестирования

### Через Postman
Импортируйте эти примеры в Postman коллекцию для удобного тестирования.

### Через скрипт
Сохраните команды в файл `test_api.sh` и выполните:
```bash
chmod +x test_api.sh
./test_api.sh
```