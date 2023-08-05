use async_trait::async_trait;
use chrono::NaiveDateTime;
use duvua_framework::errors::BotError;
use serde::{Deserialize, Serialize};
use sqlx::{Pool, Postgres, Row};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    id: i64,
    created_at: NaiveDateTime,
    updated_at: NaiveDateTime,
    coins: i32,
}

#[async_trait]
pub trait UserRepository: Sync + Send {
    async fn create(&self, id: i64) -> Result<(), BotError>;
    async fn get_by_id(&self, id: i64) -> Result<User, BotError>;
}

pub struct UserService {
    db: Pool<Postgres>,
}

impl UserService {
    pub fn new(db: Pool<Postgres>) -> Self {
        Self { db }
    }
}

#[async_trait]
impl UserRepository for UserService {
    async fn create(&self, id: i64) -> Result<(), BotError> {
        sqlx::query("INSERT INTO \"users\" (\"id\") VALUES ($1);")
            .bind(id)
            .execute(&self.db)
            .await
            .or_else(|e| {
                log::debug!("{e}");
                Err(match e {
                    sqlx::Error::Database(_) => BotError::UserAlreadyExists,
                    _ => BotError::Query,
                })
            })?;

        Ok(())
    }

    async fn get_by_id(&self, id: i64) -> Result<User, BotError> {
        let result = sqlx::query("SELECT * FROM \"users\" WHERE \"id\" = $1")
            .bind(id)
            .fetch_one(&self.db)
            .await
            .or_else(|e| {
                Err(match e {
                    sqlx::Error::Database(_) => BotError::UserNotFound,
                    _ => BotError::Query,
                })
            })?;

        Ok(User {
            id: result.get("id"),
            created_at: result.get("createdAt"),
            updated_at: result.get("updatedAt"),
            coins: result.get("coins"),
        })
    }
}
