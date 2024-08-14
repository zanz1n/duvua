-- Add up migration script here

CREATE TABLE ticket (
    id serial PRIMARY KEY,
    slug varchar(8) NOT NULL,
    created_at timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    channel_id bigint NOT NULL,
    user_id bigint NOT NULL,
    guild_id bigint NOT NULL,
    deleted_at timestamptz(3)
);

ALTER TABLE ticket ADD CONSTRAINT ticket_guild_id_fkey
    FOREIGN KEY (guild_id) REFERENCES ticket_config(guild_id)
    ON DELETE CASCADE ON UPDATE CASCADE;

CREATE UNIQUE INDEX ticket_slug_idx ON ticket(slug);
CREATE UNIQUE INDEX ticket_channel_id_idx ON ticket(channel_id);
CREATE INDEX ticket_member_idx ON ticket(user_id, guild_id);
