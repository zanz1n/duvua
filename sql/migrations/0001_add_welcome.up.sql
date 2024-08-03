-- Add up migration script here

CREATE TYPE welcometype AS ENUM ('Message', 'Image', 'Embed');

CREATE TABLE welcome (
    id BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    channel_id BIGINT,
    message VARCHAR(255) NOT NULL DEFAULT 'Seja Bem Vind@ ao servidor {{USER}}',
    kind welcometype NOT NULL DEFAULT 'Message'
);
