mod handlers;
mod repository;

use duvua_framework::{
    env::{env_param, ProcessEnv},
    handler::Handler,
};
use handlers::{component_handler::MessageComponentHandler, kiss::KissCommand, ping::PingCommand};
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

    let kiss_gifs = Arc::new(RandomStringProvider::kiss_gifs());
    let slap_gifs = Arc::new(RandomStringProvider::slap_gifs());

    let shared_handler = Arc::new(KissSharedHandler::new(kiss_gifs.clone(), slap_gifs.clone()));

    let mut handler = Handler::new(true);
    handler
        .set_component_handler(MessageComponentHandler::new(shared_handler), true)
        .add_handler(KissCommand::new(kiss_gifs))
        .add_handler(PingCommand::new());

    let intents = GatewayIntents::empty();
    let mut client = Client::builder(discord_token, intents)
        .event_handler(handler)
        .await?;

    client.start().await?;

    Ok(())
}
