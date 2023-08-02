CREATE TYPE "WelcomeType" AS ENUM ('MESSAGE', 'IMAGE', 'EMBED');

CREATE TABLE "guilds" (
    "id" BIGINT PRIMARY KEY,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "prefix" VARCHAR(4) NOT NULL DEFAULT '-',
    "enableTickets" BOOLEAN NOT NULL DEFAULT true,
    "strictMusic" BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE "welcome" (
    "id" SERIAL PRIMARY KEY,
    "guildId" BIGINT NOT NULL,
    "enabled" BOOLEAN NOT NULL DEFAULT FALSE,
    "channelId" BIGINT NOT NULL,
    "message" VARCHAR(255) NOT NULL DEFAULT 'Seja Bem Vind@ ao servidor {{USER}}',
    "type" "WelcomeType" DEFAULT 'MESSAGE'
);

CREATE UNIQUE INDEX "welcome_guildId_idx" ON "welcome"("guildId");

ALTER TABLE "welcome"
ADD CONSTRAINT "welcome_guildId_fkey" FOREIGN KEY ("guildId") REFERENCES "guilds"("id")
ON DELETE CASCADE ON UPDATE CASCADE;
