-- Rollback missing metric types added in 003_add_missing_metric_types.up.sql

DELETE FROM metric_types WHERE name IN (
  'RedisConnection',
  'PostgreSQLConnection',
  'MongoDBConnection',
  'MySQLConnection',
  'KongConnection',
  'KafkaConsumerLag'
);

