use duvua_framework::errors::BotError;
use serenity::{http::Http, model::user::User};

use crate::{redis::RedisCacheService, CacheRepository};

pub async fn get_or_store_user(
    http: impl AsRef<Http>,
    cache: impl AsRef<RedisCacheService>,
    id: u64,
) -> Result<User, BotError> {
    let key = id.to_string();

    if let Some(user) =  cache.as_ref().de_get(key.clone()).await? {
        return Ok(user);
    }

    let user = http
        .as_ref()
        .get_user(id)
        .await
        .or_else(|e| Err(BotError::Serenity(e)))?;

    cache.as_ref().ser_set_ttl(key, user.clone(), 60).await?;

    Ok(user)
}
