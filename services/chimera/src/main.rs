mod commands;
mod handlers;
mod repository;

use commands::{
    avatar::AvatarCommand, clone::CloneCommand, facts::FactsCommand, kiss::KissCommand,
    ping::PingCommand,
};
use deadpool_redis::{Config as RedisConfig, Runtime as DeadpoolRuntime};
use duvua_cache::redis::RedisCacheService;
use duvua_framework::{
    env::{env_param, ProcessEnv},
    handler::Handler,
};
use handlers::component_handler::MessageComponentHandler;
use repository::{kiss_shared::KissSharedHandler, random::RandomStringProvider};
use serenity::{prelude::GatewayIntents, Client};
use std::{error::Error, sync::Arc};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let process_env = env_param("APP_ENV", Some(ProcessEnv::None));

    if let ProcessEnv::None = process_env {
        dotenvy::dotenv().expect("Failed to open .env file, please provide environment variables");
    }
    env_logger::init();

    let discord_token: String = env_param("DISCORD_TOKEN", None);
    let redis_uri: String = env_param("REDIS_URL", None);

    let redis_client =
        RedisConfig::from_url(redis_uri).create_pool(Some(DeadpoolRuntime::Tokio1))?;
    let redis_client = Arc::new(redis_client);

    let cache_service = RedisCacheService::new(redis_client).await?;
    let cache_service = Arc::new(cache_service);

    let kiss_gifs = Arc::new(RandomStringProvider::kiss_gifs());
    let slap_gifs = Arc::new(RandomStringProvider::slap_gifs());

    let shared_handler = Arc::new(KissSharedHandler::new(kiss_gifs.clone(), slap_gifs.clone()));

    let mut handler = Handler::new(true);
    handler
        .set_component_handler(
            MessageComponentHandler::new(shared_handler, cache_service.clone()),
            true,
        )
        .add_handler(KissCommand::new(kiss_gifs, cache_service.clone()))
        .add_handler(PingCommand::new())
        .add_handler(AvatarCommand::new(cache_service.clone()))
        .add_handler(CloneCommand::new(cache_service.clone()))
        .add_handler(FactsCommand::new());

    let intents = GatewayIntents::empty();
    let mut client = Client::builder(discord_token, intents)
        .event_handler(handler)
        .await?;

    client.start().await?;

    Ok(())
}
