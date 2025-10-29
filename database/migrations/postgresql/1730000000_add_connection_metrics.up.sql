-- Add new metric types for database and service connections
INSERT INTO metric_types (name, description) VALUES 
  ('RedisConnection', 'Test Redis connection with authentication'),
  ('PostgreSQLConnection', 'Test PostgreSQL database connection with authentication'),
  ('MongoDBConnection', 'Test MongoDB database connection with authentication'),
  ('MySQLConnection', 'Test MySQL database connection with authentication'),
  ('KongConnection', 'Test Kong API Gateway connection and health');

