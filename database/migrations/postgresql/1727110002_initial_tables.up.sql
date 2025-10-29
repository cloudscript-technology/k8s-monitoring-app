CREATE TABLE projects (
	id uuid NOT NULL DEFAULT uuid_generate_v4(),
	"name" varchar(100) NOT NULL,
	"description" text NOT NULL,
	CONSTRAINT projects_pk PRIMARY KEY (id),
	CONSTRAINT projects_name_uk UNIQUE (name)
);

CREATE TABLE applications (
	id uuid NOT NULL DEFAULT uuid_generate_v4(),
	project_id uuid NOT NULL,
	"name" varchar(100) NOT NULL,
	"description" text NOT NULL,
	"namespace" varchar(100) NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT now(),
	"updated_at" timestamp NOT NULL DEFAULT now(),
	CONSTRAINT applications_pk PRIMARY KEY (id),
	CONSTRAINT applications_name_uk UNIQUE (name, namespace),
	CONSTRAINT applications_project_fk FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE metric_types (
	id uuid NOT NULL DEFAULT uuid_generate_v4(),
	"name" text NOT NULL,
	"description" text NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT now(),
	"updated_at" timestamp NOT NULL DEFAULT now(),
	CONSTRAINT metric_types_pk PRIMARY KEY (id),
	CONSTRAINT metric_types_name_uk UNIQUE (name)
);

INSERT INTO metric_types (name, description) VALUES ('PodActiveNodes', 'Number of active nodes in the pod');
INSERT INTO metric_types (name, description) VALUES ('PodStatus', 'Status of the pod');
INSERT INTO metric_types (name, description) VALUES ('PodMemoryUsage', 'Memory usage of the pod');
INSERT INTO metric_types (name, description) VALUES ('PodCpuUsage', 'CPU usage of the pod');
INSERT INTO metric_types (name, description) VALUES ('PvcUsage', 'Usage of the PVC');
INSERT INTO metric_types (name, description) VALUES ('HealthCheck', 'Health check of the pod');

CREATE TABLE application_metrics (
	id uuid NOT NULL DEFAULT uuid_generate_v4(),
	application_id uuid NOT NULL,
	"type_id" uuid NOT NULL,
	"configuration" jsonb NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT now(),
	"updated_at" timestamp NOT NULL DEFAULT now(),
	CONSTRAINT application_metrics_pk PRIMARY KEY (id),
	CONSTRAINT application_metrics_application_fk FOREIGN KEY (application_id) REFERENCES applications(id),
	CONSTRAINT application_metrics_type_fk FOREIGN KEY (type_id) REFERENCES metric_types(id)
);

CREATE TABLE application_metric_values (
	id uuid NOT NULL DEFAULT uuid_generate_v4(),
	application_metric_id uuid NOT NULL,
	"value" jsonb NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT now(),
	"updated_at" timestamp NOT NULL DEFAULT now(),
	CONSTRAINT application_metric_values_pk PRIMARY KEY (id),
	CONSTRAINT application_metric_values_application_metric_fk FOREIGN KEY (application_metric_id) REFERENCES application_metrics(id)
);