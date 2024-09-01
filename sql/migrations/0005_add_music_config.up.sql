-- Add up migration script here

CREATE TYPE musicpermission AS ENUM ('All', 'DJ', 'Adm');

CREATE TABLE music_config (
    guild_id bigint PRIMARY KEY,
    created_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled boolean NOT NULL DEFAULT TRUE,
    play_mode musicpermission NOT NULL DEFAULT 'All',
    control_mode musicpermission NOT NULL DEFAULT 'DJ',
    dj_role bigint
);

CREATE TRIGGER update_music_config_updated_at
   BEFORE UPDATE ON music_config FOR EACH ROW
   EXECUTE PROCEDURE update_modified_column();
