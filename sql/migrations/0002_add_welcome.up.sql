-- Add up migration script here

CREATE TYPE welcometype AS ENUM ('Message', 'Image', 'Embed');

CREATE TABLE welcome (
    id bigint PRIMARY KEY,
    created_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled boolean NOT NULL DEFAULT FALSE,
    channel_id bigint,
    message varchar(255) NOT NULL DEFAULT 'Seja Bem Vind@ ao servidor {{USER}}',
    kind welcometype NOT NULL DEFAULT 'Message'
);

CREATE TRIGGER update_welcome_updated_at
   BEFORE UPDATE ON welcome FOR EACH ROW
   EXECUTE PROCEDURE update_modified_column();
