-- Add down migration script here

DROP TRIGGER IF EXISTS update_welcome_updated_at ON welcome;

DROP TABLE IF EXISTS welcome;
DROP TYPE IF EXISTS welcometype;
