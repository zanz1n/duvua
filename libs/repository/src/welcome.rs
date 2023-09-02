use async_trait::async_trait;
use chrono::{NaiveDateTime, Utc};
use duvua_framework::errors::BotError;
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgRow, FromRow, Pool, Postgres, Row};
use std::{fmt::Display, io::ErrorKind, str::FromStr};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "welcometype", rename_all = "UPPERCASE")]
#[serde(rename_all = "UPPERCASE")]
pub enum WelcomeType {
    Message,
    Image,
    Embed,
}

impl Display for WelcomeType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(match self {
            Self::Message => "MESSAGE",
            Self::Image => "IMAGE",
            Self::Embed => "EMBED",
        })
    }
}

impl FromStr for WelcomeType {
    type Err = std::io::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "MESSAGE" => Ok(Self::Message),
            "IMAGE" => Ok(Self::Image),
            "EMBED" => Ok(Self::Embed),
            _ => Err(std::io::Error::new(
                ErrorKind::InvalidData,
                "WelcomeType enumerator must be MESSAGE | IMAGE | EMBED",
            )),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Welcome {
    pub id: i64,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub enabled: bool,
    pub channel_id: Option<i64>,
    pub message: String,
    pub kind: WelcomeType,
}

impl Default for Welcome {
    fn default() -> Self {
        let now = Utc::now().naive_utc();

        Self {
            id: Default::default(),
            created_at: now,
            updated_at: now,
            enabled: false,
            channel_id: None,
            message: "Seja Bem Vind@ ao servidor {{USER}}".to_owned(),
            kind: WelcomeType::Message,
        }
    }
}

impl FromRow<'_, PgRow> for Welcome {
    fn from_row(row: &'_ PgRow) -> Result<Self, sqlx::Error> {
        let welcome = Self {
            id: row.try_get("id")?,
            created_at: row.try_get("createdAt")?,
            updated_at: row.try_get("updatedAt")?,
            enabled: row.try_get("enabled")?,
            channel_id: row.try_get("channelId")?,
            message: row.try_get("message")?,
            kind: row.try_get("type")?,
        };

        Ok(welcome)
    }
}

#[async_trait]
pub trait WelcomeRepository: Sync + Send {
    async fn get_by_id(&self, id: i64) -> Result<Option<Welcome>, BotError>;
    async fn exists(&self, id: i64) -> Result<bool, BotError>;
    async fn update_enabled(&self, id: i64, enabled: bool) -> Result<(), BotError>;
    async fn update_channel_id(&self, id: i64, channel_id: Option<i64>) -> Result<(), BotError>;
    async fn update_message(
        &self,
        id: i64,
        message: String,
        kind: WelcomeType,
    ) -> Result<(), BotError>;
    async fn create(&self, data: Welcome) -> Result<Welcome, BotError>;
    async fn create_default(&self, id: i64) -> Result<Welcome, BotError>;
    async fn delete_by_id(&self, id: i64) -> Result<bool, BotError>;
}

pub struct WelcomeService {
    db: Pool<Postgres>,
}

impl WelcomeService {
    pub fn new(db: Pool<Postgres>) -> Self {
        Self { db }
    }
}

#[async_trait]
impl WelcomeRepository for WelcomeService {
    async fn get_by_id(&self, id: i64) -> Result<Option<Welcome>, BotError> {
        let row = sqlx::query_as("SELECT * FROM \"welcome\" WHERE \"id\" = $1;")
            .bind(id)
            .fetch_one(&self.db)
            .await;

        let row = match row {
            Ok(v) => Some(v),
            Err(e) => match e {
                sqlx::Error::RowNotFound => None,
                _ => {
                    log::error!(target: "welcome_service.get_by_id", "{e}");
                    return Err(BotError::PostgresError);
                }
            },
        };

        Ok(row)
    }

    async fn exists(&self, id: i64) -> Result<bool, BotError> {
        let row = sqlx::query(r#"SELECT "id" FROM "welcome" WHERE "id" = $1;"#)
            .bind(id)
            .fetch_one(&self.db)
            .await;

        let row = match row {
            Ok(v) => v,
            Err(e) => match e {
                sqlx::Error::RowNotFound => return Ok(false),
                _ => {
                    log::error!(target: "welcome_service.exists", "{e}");
                    return Err(BotError::PostgresError);
                }
            },
        };

        let row_id: i64 = row.try_get("id").or_else(|e| {
            log::error!(target: "welcome_service.exists", "Failed to get \"row\".\"id\": {e}");
            Err(BotError::PostgresError)
        })?;

        Ok(row_id == id)
    }

    async fn update_enabled(&self, id: i64, enabled: bool) -> Result<(), BotError> {
        sqlx::query(
            r#"UPDATE "welcome" SET
            "updatedAt" = CURRENT_TIMESTAMP,
            "enabled" = $1
            WHERE "id" = $2;"#,
        )
        .bind(enabled)
        .bind(id)
        .execute(&self.db)
        .await
        .or_else(|e| {
            log::error!(target: "welcome_service.update_enabled", "{e}");
            Err(BotError::PostgresError)
        })?;

        Ok(())
    }

    async fn update_channel_id(&self, id: i64, channel_id: Option<i64>) -> Result<(), BotError> {
        sqlx::query(
            r#"UPDATE "welcome" SET
            "updatedAt" = CURRENT_TIMESTAMP,
            "channelId" = $1
            WHERE "id" = $2;"#,
        )
        .bind(channel_id)
        .bind(id)
        .execute(&self.db)
        .await
        .or_else(|e| {
            log::error!(target: "welcome_service.update_channel_id", "{e}");
            Err(BotError::PostgresError)
        })?;

        Ok(())
    }

    async fn update_message(
        &self,
        id: i64,
        message: String,
        kind: WelcomeType,
    ) -> Result<(), BotError> {
        sqlx::query(
            r#"UPDATE "welcome" SET
            "updatedAt" = CURRENT_TIMESTAMP,
            "message" = $1,
            "type" = $2
            WHERE "id" = $3;"#,
        )
        .bind(message)
        .bind(kind)
        .bind(id)
        .execute(&self.db)
        .await
        .or_else(|e| {
            log::error!(target: "welcome_service.update_message", "{e}");
            Err(BotError::PostgresError)
        })?;

        Ok(())
    }

    async fn create(&self, data: Welcome) -> Result<Welcome, BotError> {
        let row = sqlx::query_as(
            r#"INSERT INTO "welcome"
            ("id", "createdAt", "updatedAt", "enabled", "channelId", "message", "type")
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING *;"#,
        )
        .bind(data.id)
        .bind(data.created_at)
        .bind(data.updated_at)
        .bind(data.enabled)
        .bind(data.channel_id)
        .bind(data.message)
        .bind(data.kind)
        .fetch_one(&self.db)
        .await
        .or_else(|e| {
            log::error!(target: "welcome_service.create", "{e}");
            Err(BotError::PostgresError)
        })?;

        Ok(row)
    }

    async fn create_default(&self, id: i64) -> Result<Welcome, BotError> {
        let row = sqlx::query_as(r#"INSERT INTO "welcome" ("id") VALUES ($1) RETURNING *;"#)
            .bind(id)
            .fetch_one(&self.db)
            .await
            .or_else(|e| {
                log::error!(target: "welcome_service.create_default", "{e}");
                Err(BotError::PostgresError)
            })?;

        Ok(row)
    }

    async fn delete_by_id(&self, id: i64) -> Result<bool, BotError> {
        let result = sqlx::query(r#"DELETE FROM "welcome" WHERE "id" = $1;"#)
            .bind(id)
            .execute(&self.db)
            .await
            .or_else(|e| {
                log::error!(target: "welcome_service.delete_by_id", "{e}");
                Err(BotError::PostgresError)
            })?;

        Ok(result.rows_affected() == 1)
    }
}
