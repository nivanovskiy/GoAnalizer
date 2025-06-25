# Performance Analyzer - Go Application

## Overview

This is a Go-based performance testing analysis application that provides a REST API for managing performance test workflows. The system processes project files, analyzes them using an AI model, stores test results, and generates comprehensive performance analysis reports.

## System Architecture

The application follows a clean architecture pattern with distinct layers:

- **Handlers Layer**: HTTP request handling and routing using Gin framework
- **Services Layer**: Business logic including AI client integration and background analysis processing
- **Database Layer**: PostgreSQL data persistence with connection pooling via pgx/v5
- **Models Layer**: Data structures and domain entities
- **Configuration Layer**: Environment-based configuration management

## Key Components

### API Endpoints
- `POST /initAnalize/{tenant}/{repo}/{uuid}` - Initialize analysis pipeline for a project
- `POST /sendFile/{uuid}` - Submit project files for analysis
- `POST /sendResults/{uuid}` - Submit performance test results
- `GET /getAnalizeResults/{uuid}` - Retrieve analysis results

### Background Processing
- Asynchronous analysis engine that processes files and test results
- Queue-based system for handling analysis requests
- AI model integration for expert-level performance analysis

### Database Schema
- **projects**: Store project metadata and status
- **project_files**: Store uploaded files and their individual analyses
- **test_results**: Store performance test metrics and results
- **analysis_results**: Store final AI-generated analysis reports

### AI Integration
- REST API client for external AI model communication
- Configurable endpoint (default: localhost:1234)
- Structured prompts for performance analysis expertise
- JSON response parsing and storage

## Data Flow

1. **Initialization**: Client calls `/initAnalize` to start a new analysis session
2. **File Submission**: Multiple calls to `/sendFile` upload project files
3. **Test Results**: Client submits performance metrics via `/sendResults`
4. **Background Analysis**: System triggers AI analysis of all collected data
5. **Result Retrieval**: Client polls `/getAnalizeResults` for completed analysis

## External Dependencies

- **PostgreSQL**: Primary data storage with JSONB support for flexible schema
- **AI Model Service**: External REST API for performance analysis (localhost:1234)
- **Gin Framework**: HTTP router and middleware
- **pgx/v5**: PostgreSQL driver with connection pooling

## Deployment Strategy

The application is configured for Replit deployment with:
- Go module initialization and dependency management
- Automatic server startup on port 5000
- Environment variable configuration support
- Database migration system for schema management

## User Preferences

Preferred communication style: Simple, everyday language.

## Changelog

Changelog:
- June 25, 2025: Initial setup and complete implementation
- Implemented all 4 required API endpoints with PostgreSQL storage
- Added AI model integration with mock fallback responses
- Background analysis processing working correctly
- Database migrations and connection pool configured
- Server successfully running on port 5000 with API documentation