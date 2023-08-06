mod handlers;
mod repository;

use crate::{
    handlers::{ticket::TicketCommandHandler, ticket_shared::TicketSharedHandler},
    repository::{
        guild::{GuildRepository, GuildService},
        ticket::{TicketRepository, TicketService},
    },
};
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
        dotenvy::dotenv()
            .expect("Failed to open .env file, please provide environment variables or the file");
    }
    env_logger::init();

    let mongo_uri: String = env_param("MONGO_URI", None);
    let discord_token: String = env_param("DISCORD_TOKEN", None);

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

    let guild_repo: Arc<dyn GuildRepository> = Arc::new(GuildService::new(db.collection("guilds")));
    let ticket_repo: Arc<dyn TicketRepository> =
        Arc::new(TicketService::new(db.collection("tickets")));

    let mut handler = Handler::new(true);
    let ticket_shared_handler = Arc::new(TicketSharedHandler::new(ticket_repo));

    handler.add_handler(TicketCommandHandler::new(
        guild_repo,
        ticket_shared_handler.clone(),
    ));

    let intents = GatewayIntents::empty();
    let mut client = Client::builder(discord_token, intents)
        .event_handler(handler)
        .await?;

    client.start().await?;

    Ok(())
}
