use duvua_framework::errors::BotError;
use serenity::{
    http::Http,
    model::{prelude::PartialGuild, user::User},
};

use crate::{redis::RedisCacheService, CacheRepository};

pub async fn get_or_store_user(
    http: impl AsRef<Http>,
    cache: impl AsRef<RedisCacheService>,
    id: u64,
) -> Result<User, BotError> {
    let key = id.to_string();

    if let Some(user) = cache.as_ref().de_get(key.clone()).await? {
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

pub async fn get_or_store_guild(
    http: impl AsRef<Http>,
    cache: impl AsRef<RedisCacheService>,
    id: u64,
) -> Result<PartialGuild, BotError> {
    let key = id.to_string();

    if let Some(guild) = cache.as_ref().de_get(key.clone()).await? {
        return Ok(guild);
    }

    let guild = http
        .as_ref()
        .get_guild(id)
        .await
        .or_else(|e| Err(BotError::Serenity(e)))?;

    cache.as_ref().ser_set_ttl(key, guild.clone(), 60).await?;

    Ok(guild)
}
