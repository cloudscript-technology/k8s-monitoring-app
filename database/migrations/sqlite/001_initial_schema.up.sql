-- SQLite version of the complete database schema
-- Combines all PostgreSQL migrations into a single SQLite-compatible schema

-- Create projects table
CREATE TABLE projects (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL
);

-- Create applications table
CREATE TABLE applications (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    project_id TEXT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    UNIQUE(name, namespace)
);

-- Create metric_types table
CREATE TABLE metric_types (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default metric types
INSERT INTO metric_types (name, description) VALUES ('PodActiveNodes', 'Number of active nodes in the pod');
INSERT INTO metric_types (name, description) VALUES ('PodStatus', 'Status of the pod');
INSERT INTO metric_types (name, description) VALUES ('PodMemoryUsage', 'Memory usage of the pod');
INSERT INTO metric_types (name, description) VALUES ('PodCpuUsage', 'CPU usage of the pod');
INSERT INTO metric_types (name, description) VALUES ('PvcUsage', 'Usage of the PVC');
INSERT INTO metric_types (name, description) VALUES ('HealthCheck', 'Health check of the pod');
INSERT INTO metric_types (name, description) VALUES ('IngressCertificate', 'Ingress certificate expiration monitoring');

-- Create application_metrics table
CREATE TABLE application_metrics (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    application_id TEXT NOT NULL,
    type_id TEXT NOT NULL,
    configuration TEXT NOT NULL, -- JSON stored as TEXT in SQLite
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id),
    FOREIGN KEY (type_id) REFERENCES metric_types(id)
);

-- Create application_metric_values table
CREATE TABLE application_metric_values (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    application_metric_id TEXT NOT NULL,
    value TEXT NOT NULL, -- JSON stored as TEXT in SQLite
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_metric_id) REFERENCES application_metrics(id)
);

-- Create sessions table for OAuth authentication
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    user_picture TEXT, -- Added from later migration
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expiry DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_applications_project_id ON applications(project_id);
CREATE INDEX idx_applications_name_namespace ON applications(name, namespace);
CREATE INDEX idx_application_metrics_application_id ON application_metrics(application_id);
CREATE INDEX idx_application_metrics_type_id ON application_metrics(type_id);
CREATE INDEX idx_application_metric_values_metric_id ON application_metric_values(application_metric_id);
CREATE INDEX idx_application_metric_values_created_at ON application_metric_values(created_at);
CREATE INDEX idx_sessions_user_email ON sessions(user_email);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Create triggers to update updated_at timestamps
CREATE TRIGGER update_applications_updated_at 
    AFTER UPDATE ON applications
    FOR EACH ROW
    BEGIN
        UPDATE applications SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER update_metric_types_updated_at 
    AFTER UPDATE ON metric_types
    FOR EACH ROW
    BEGIN
        UPDATE metric_types SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER update_application_metrics_updated_at 
    AFTER UPDATE ON application_metrics
    FOR EACH ROW
    BEGIN
        UPDATE application_metrics SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER update_application_metric_values_updated_at 
    AFTER UPDATE ON application_metric_values
    FOR EACH ROW
    BEGIN
        UPDATE application_metric_values SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;