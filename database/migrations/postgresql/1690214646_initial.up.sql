CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
      NEW.updated_at = now(); 
      RETURN NEW;
END;
$$ language 'plpgsql';

SET TIME ZONE 'UTC';
