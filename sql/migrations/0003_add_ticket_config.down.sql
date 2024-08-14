-- Add down migration script here

DROP TRIGGER IF EXISTS update_ticket_config_updated_at ON ticket;

DROP TABLE IF EXISTS ticket_config;
