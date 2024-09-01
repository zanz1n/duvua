-- Add down migration script here

DROP TRIGGER IF EXISTS update_music_config_updated_at ON music_config;

DROP TABLE IF EXISTS welcome;
DROP TYPE IF EXISTS musicpermission;
