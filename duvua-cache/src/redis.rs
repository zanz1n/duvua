use crate::{de, ser, CacheRepository};
use async_trait::async_trait;
use deadpool_redis::{
    redis::{self, AsyncCommands, Expiry},
    Connection, Pool,
};
use duvua_framework::errors::BotError;
use serde::{de::DeserializeOwned, Serialize};
use std::{
    io::{self, Error},
    sync::Arc,
};

#[derive(Clone)]
pub struct RedisCacheService {
    client: Arc<Pool>,
}

impl RedisCacheService {
    pub async fn new(client: Arc<Pool>) -> Result<Self, Error> {
        let mut conn = client
            .get()
            .await
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        redis::cmd("PING")
            .query_async::<_, ()>(&mut conn)
            .await
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        Ok(Self { client })
    }

    pub async fn get_conn(&self) -> Result<Connection, BotError> {
        match self.client.get().await {
            Ok(c) => Ok(c),
            Err(e) => {
                log::error!("Failed to get_conn(): {}", e);
                Err(BotError::RedisError)
            }
        }
    }
}

#[async_trait]
impl CacheRepository for RedisCacheService {
    async fn get(&self, key: String) -> Result<Option<String>, BotError> {
        let mut conn = self.get_conn().await?;

        let value: Option<String> = conn.get(key).await.or_else(|e| {
            log::error!("Failed to get(): {}", e);
            Err(BotError::RedisError)
        })?;

        Ok(value)
    }

    async fn get_ttl(&self, key: String, ttl: usize) -> Result<Option<String>, BotError> {
        let mut conn = self.get_conn().await?;

        let value: Option<String> = conn.get_ex(key, Expiry::EX(ttl)).await.or_else(|e| {
            log::error!("Failed to get_ttl(): {}", e);
            Err(BotError::RedisError)
        })?;

        Ok(value)
    }

    async fn set(&self, key: String, value: String) -> Result<(), BotError> {
        let mut conn = self.get_conn().await?;

        conn.set(key, value).await.or_else(|e| {
            log::error!("Failed to set(): {}", e);
            Err(BotError::RedisError)
        })?;

        Ok(())
    }

    async fn set_ttl(&self, key: String, value: String, ttl: usize) -> Result<(), BotError> {
        let mut conn = self.get_conn().await?;

        conn.set_ex(key, value, ttl).await.or_else(|e| {
            log::error!("Failed to set_ttl(): {}", e);
            Err(BotError::RedisError)
        })?;

        Ok(())
    }

    async fn del(&self, key: String) -> Result<(), BotError> {
        let mut conn = self.get_conn().await?;

        conn.del(key).await.or_else(|e| {
            log::error!("Failed to set_ttl(): {}", e);
            Err(BotError::RedisError)
        })?;

        Ok(())
    }

    async fn de_get<T: Sized + DeserializeOwned>(
        &self,
        key: String,
    ) -> Result<Option<T>, BotError> {
        let cache = self.get(key).await?;

        Ok(match cache {
            Some(cache) => Some(de(&cache)?),
            None => None,
        })
    }

    async fn de_get_ttl<T: Sized + DeserializeOwned>(
        &self,
        key: String,
        ttl: usize,
    ) -> Result<Option<T>, BotError> {
        let cache = self.get_ttl(key, ttl).await?;

        Ok(match cache {
            Some(cache) => Some(de(&cache)?),
            None => None,
        })
    }

    async fn ser_set<T: Sized + Serialize + Send>(
        &self,
        key: String,
        value: T,
    ) -> Result<(), BotError> {
        let value = ser(value)?;
        self.set(key, value).await
    }

    async fn ser_set_ttl<T: Sized + Serialize + Send>(
        &self,
        key: String,
        value: T,
        ttl: usize,
    ) -> Result<(), BotError> {
        let value = ser(value)?;
        self.set_ttl(key, value, ttl).await
    }
}
