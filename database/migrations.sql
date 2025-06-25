-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create projects table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    tenant VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    uuid UUID UNIQUE NOT NULL,
    language VARCHAR(100),
    testing_tool VARCHAR(100),
    project_info JSONB,
    files_count INTEGER DEFAULT 0,
    received_files_count INTEGER DEFAULT 0,
    has_test_results BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'initialized',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create project_files table
CREATE TABLE IF NOT EXISTS project_files (
    id SERIAL PRIMARY KEY,
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    filename VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    file_analysis JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_uuid, filename)
);

-- Create test_results table
CREATE TABLE IF NOT EXISTS test_results (
    id SERIAL PRIMARY KEY,
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    response_time_p95 JSONB,
    response_time_p99 JSONB,
    successful_calls INTEGER,
    failed_calls INTEGER,
    nonfunctional_requirements JSONB,
    raw_results JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create analysis_results table
CREATE TABLE IF NOT EXISTS analysis_results (
    id SERIAL PRIMARY KEY,
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    final_analysis JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_projects_uuid ON projects(uuid);
CREATE INDEX IF NOT EXISTS idx_projects_tenant_repo ON projects(tenant, repo);
CREATE INDEX IF NOT EXISTS idx_project_files_uuid ON project_files(project_uuid);
CREATE INDEX IF NOT EXISTS idx_test_results_uuid ON test_results(project_uuid);
CREATE INDEX IF NOT EXISTS idx_analysis_results_uuid ON analysis_results(project_uuid);
CREATE INDEX IF NOT EXISTS idx_analysis_results_status ON analysis_results(status);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
