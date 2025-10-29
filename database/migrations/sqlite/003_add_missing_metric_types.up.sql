-- Add missing metric types in SQLite to match PostgreSQL migrations
-- This migration aligns the metric_types with connection and Kafka lag metrics

-- Database and Service Connection metric types
INSERT INTO metric_types (name, description) VALUES ('RedisConnection', 'Test Redis connection with authentication');
INSERT INTO metric_types (name, description) VALUES ('PostgreSQLConnection', 'Test PostgreSQL database connection with authentication');
INSERT INTO metric_types (name, description) VALUES ('MongoDBConnection', 'Test MongoDB database connection with authentication');
INSERT INTO metric_types (name, description) VALUES ('MySQLConnection', 'Test MySQL database connection with authentication');
INSERT INTO metric_types (name, description) VALUES ('KongConnection', 'Test Kong API Gateway connection and health');

-- Kafka consumer lag metric type
INSERT INTO metric_types (name, description) VALUES ('KafkaConsumerLag', 'Monitor Kafka consumer lag for topics and consumer groups');

-- Note: We intentionally do not remove legacy placeholders like 'ConnectionMetrics' or 'KafkaLag'
-- to avoid breaking existing references. They can be cleaned up later if unused.

