-- Remove connection metric types
DELETE FROM metric_types WHERE name IN (
  'RedisConnection',
  'PostgreSQLConnection',
  'MongoDBConnection',
  'MySQLConnection',
  'KongConnection'
);

