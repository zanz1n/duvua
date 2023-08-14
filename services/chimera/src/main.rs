mod handlers;

use duvua_framework::{
    env::{env_param, ProcessEnv},
    handler::Handler,
};
use handlers::{component_handler::MessageComponentHandler, ping::PingCommand};
use serenity::{prelude::GatewayIntents, Client};
use std::error::Error;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let process_env = env_param("APP_ENV", Some(ProcessEnv::None));

    if let ProcessEnv::None = process_env {
        dotenvy::dotenv()
            .expect("Failed to open .env file, please provide environment variables or the file");
    }
    env_logger::init();

    let discord_token: String = env_param("DISCORD_TOKEN", None);

    let mut handler = Handler::new(true);
    handler
        .set_component_handler(MessageComponentHandler::new(), true)
        .add_handler(PingCommand::new());

    let intents = GatewayIntents::empty();
    let mut client = Client::builder(discord_token, intents)
        .event_handler(handler)
        .await?;

    client.start().await?;

    Ok(())
}
