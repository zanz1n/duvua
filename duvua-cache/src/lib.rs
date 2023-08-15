pub mod redis;

use async_trait::async_trait;
use duvua_framework::errors::BotError;
use serde::{de::DeserializeOwned, Serialize};

pub(crate) fn de<T: Sized + DeserializeOwned>(value: &str) -> Result<T, BotError> {
    match serde_json::from_str::<T>(value) {
        Ok(v) => Ok(v),
        Err(e) => {
            log::error!("Failed do deserialize cache: {e}");
            Err(BotError::CacheDeserializeError)
        }
    }
}

pub(crate) fn ser<T: Sized + Serialize>(value: T) -> Result<String, BotError> {
    match serde_json::to_string(&value) {
        Ok(v) => Ok(v),
        Err(e) => {
            log::error!("Failed do serialize cache: {e}");
            Err(BotError::CacheDeserializeError)
        }
    }
}

#[async_trait]
pub trait CacheRepository: Sync + Send {
    async fn get(&self, key: String) -> Result<Option<String>, BotError>;
    async fn get_ttl(&self, key: String, ttl: usize) -> Result<Option<String>, BotError>;
    async fn set(&self, key: String, value: String) -> Result<(), BotError>;
    async fn set_ttl(&self, key: String, value: String, ttl: usize) -> Result<(), BotError>;
    async fn del(&self, key: String) -> Result<(), BotError>;

    async fn de_get<T: Sized + DeserializeOwned>(&self, key: String)
        -> Result<Option<T>, BotError>;
    async fn de_get_ttl<T: Sized + DeserializeOwned>(
        &self,
        key: String,
        ttl: usize,
    ) -> Result<Option<T>, BotError>;
    async fn ser_set<T: Sized + Serialize + Send>(
        &self,
        key: String,
        value: T,
    ) -> Result<(), BotError>;
    async fn ser_set_ttl<T: Sized + Serialize + Send>(
        &self,
        key: String,
        value: T,
        ttl: usize,
    ) -> Result<(), BotError>;
}
