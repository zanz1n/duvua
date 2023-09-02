use async_trait::async_trait;
use chrono::NaiveDateTime;
use duvua_framework::errors::BotError;
use sqlx::{postgres::PgRow, FromRow, Pool, Postgres, Row};

pub struct Guild {
    pub id: i64,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub prefix: String,
    pub strict_music: bool,
}

impl FromRow<'_, PgRow> for Guild {
    fn from_row(row: &'_ PgRow) -> Result<Self, sqlx::Error> {
        let id = row.try_get("id")?;
        let created_at = row.try_get("createdAt")?;
        let updated_at = row.try_get("updatedAt")?;
        let prefix = row.try_get("prefix")?;
        let strict_music = row.try_get("strictMusic")?;

        let welcome = Self {
            id,
            created_at,
            updated_at,
            prefix,
            strict_music,
        };

        Ok(welcome)
    }
}

#[async_trait]
pub trait GuildRepository {
    async fn exists(&self, id: i64) -> Result<bool, BotError>;
    async fn create_default(&self, id: i64) -> Result<Guild, BotError>;
    async fn delete_by_id(&self, id: i64) -> Result<bool, BotError>;
}

pub struct GuildService {
    db: Pool<Postgres>,
}

impl GuildService {
    pub fn new(db: Pool<Postgres>) -> Self {
        Self { db }
    }
}

#[async_trait]
impl GuildRepository for GuildService {
    async fn exists(&self, id: i64) -> Result<bool, BotError> {
        let row = sqlx::query(r#"SELECT "id" FROM "guilds" WHERE "id" = $1;"#)
            .bind(id)
            .fetch_one(&self.db)
            .await;

        let row = match row {
            Ok(v) => v,
            Err(e) => {
                log::error!(target: "guild_service.exists", "{e}");
                match e {
                    sqlx::Error::RowNotFound => return Ok(false),
                    _ => return Err(BotError::PostgresError),
                }
            }
        };

        let row_id: i64 = row.try_get("id").or_else(|e| {
            log::error!(target: "guild_service.exists", "{e}");
            Err(BotError::PostgresError)
        })?;

        Ok(row_id == id)
    }

    async fn create_default(&self, id: i64) -> Result<Guild, BotError> {
        let row = sqlx::query_as(r#"INSERT INTO "guilds" ("id") VALUES ($1) RETURNING *;"#)
            .bind(id)
            .fetch_one(&self.db)
            .await
            .or_else(|e| {
                println!("{e}");
                log::error!(target: "guild_service.create_default", "{e}");
                Err(BotError::PostgresError)
            })?;

        Ok(row)
    }

    async fn delete_by_id(&self, id: i64) -> Result<bool, BotError> {
        let result = sqlx::query(r#"DELETE FROM "guilds" WHERE "id" = $1;"#)
            .bind(id)
            .execute(&self.db)
            .await
            .or_else(|e| {
                log::error!(target: "guild_service.delete_by_id", "{e}");
                Err(BotError::PostgresError)
            })?;

        Ok(result.rows_affected() == 1)
    }
}
