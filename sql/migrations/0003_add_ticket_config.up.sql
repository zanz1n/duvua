-- Add up migration script here

CREATE TABLE ticket_config (
    guild_id bigint PRIMARY KEY,
    created_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled boolean NOT NULL DEFAULT FALSE,
    allow_multiple boolean NOT NULL DEFAULT TRUE,
    channel_name varchar(24),
    channel_category_id bigint,
    logs_channel_id bigint
);

CREATE TRIGGER update_ticket_config_updated_at
   BEFORE UPDATE ON ticket_config FOR EACH ROW
   EXECUTE PROCEDURE update_modified_column();
