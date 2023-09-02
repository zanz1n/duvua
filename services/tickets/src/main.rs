mod handlers;
mod repository;

use crate::{
    handlers::{
        component_handler::MessageComponentHandler, ticket::TicketCommandHandler,
        ticketadmin::TicketAdminCommandHandler,
    },
    repository::{guild::GuildService, ticket::TicketService, ticket_shared::TicketSharedHandler},
};
use deadpool_redis::{Config as RedisConfig, Runtime as DeadpoolRuntime};
use duvua_framework::{
    env::{env_param, ProcessEnv},
    handler::Handler,
};
use mongodb::{bson::doc, options::ClientOptions, Client as MongoDbClient};
use serenity::{prelude::GatewayIntents, Client};
use std::{error::Error, sync::Arc, time::Instant};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let process_env = env_param("APP_ENV", Some(ProcessEnv::None));

    if let ProcessEnv::None = process_env {
        dotenvy::dotenv().expect("Failed to open .env file, please provide environment variables");
    }
    env_logger::init();
    let process_env = env_param("APP_ENV", None);

    let mongo_uri: String = env_param("MONGO_URI", None);
    let discord_token: String = env_param("DISCORD_TOKEN", None);
    let redis_uri: String = env_param("REDIS_URL", None);

    let redis_client =
        RedisConfig::from_url(redis_uri).create_pool(Some(DeadpoolRuntime::Tokio1))?;
    let redis_client = Arc::new(redis_client);

    let options = ClientOptions::parse(mongo_uri).await?;

    let start = Instant::now();

    let client = MongoDbClient::with_options(options)?;
    _ = client
        .database("admin")
        .run_command(doc! {"ping": 1}, None)
        .await?;

    log::info!(
        "Connected to mongo, handshake took {}ms",
        (Instant::now() - start).as_millis()
    );

    let db = client.database("duvua-tickets");

    let guild_repo = Arc::new(GuildService::new(db.collection("guilds")));
    let ticket_repo = Arc::new(TicketService::new(db.collection("tickets")));

    let mut handler = Handler::new(if let ProcessEnv::Production = process_env {
        Some(redis_client.clone())
    } else {
        None
    });

    let ticket_shared = Arc::new(TicketSharedHandler::new(ticket_repo.clone()));

    handler
        .add_handler(TicketCommandHandler::new(
            guild_repo.clone(),
            ticket_repo.clone(),
            ticket_shared.clone(),
        ))
        .add_handler(TicketAdminCommandHandler::new(ticket_repo, guild_repo))
        .set_component_handler(MessageComponentHandler::new(ticket_shared), true);

    let intents = GatewayIntents::empty();
    let mut client = Client::builder(discord_token, intents)
        .event_handler(handler)
        .await?;

    client.start().await?;

    Ok(())
}
