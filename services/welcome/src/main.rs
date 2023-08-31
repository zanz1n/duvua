mod handlers;

use deadpool_redis::{Config as RedisConfig, Runtime as DeadpoolRuntime};
use duvua_framework::{
    env::{env_param, ProcessEnv},
    handler::Handler,
};
use duvua_repository::{sqlx::PgPool, welcome::WelcomeService};
use handlers::serenity_handler::SerenityHandler;
use serenity::{prelude::GatewayIntents, Client};
use std::sync::Arc;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let process_env = env_param("APP_ENV", Some(ProcessEnv::None));

    if let ProcessEnv::None = process_env {
        dotenvy::dotenv().expect("Failed to open .env file, please provide environment variables");
    }
    env_logger::init();

    let discord_token: String = env_param("DISCORD_TOKEN", None);
    let database_url: String = env_param("DATABASE_URL", None);
    let redis_uri: String = env_param("REDIS_URL", None);

    let redis_client =
        RedisConfig::from_url(redis_uri).create_pool(Some(DeadpoolRuntime::Tokio1))?;
    let redis_client = Arc::new(redis_client);

    let db_pool = PgPool::connect(&database_url).await?;
    let welcome_service = WelcomeService::new(db_pool);

    let handler = Handler::new(if let ProcessEnv::Production = process_env {
        Some(redis_client.clone())
    } else {
        None
    });

    let event_handler = SerenityHandler::new(handler);

    let intents = GatewayIntents::empty().union(GatewayIntents::GUILD_MEMBERS);
    let mut client = Client::builder(discord_token, intents)
        .event_handler(event_handler)
        .await?;

    client.start().await?;

    Ok(())
}
