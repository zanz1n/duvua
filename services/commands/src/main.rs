mod messaging;

use crate::messaging::{extract_payload_data, Payload};
use messaging::{spawn_daemon, SubClient};
use serde_json::Value;
use serenity::{http::Http, model::prelude::command::Command};
use std::{env, process, sync::Arc, time::Duration};

#[inline]
fn display_value(value: Option<&Value>) -> String {
    match value.unwrap_or(&Value::Null) {
        Value::String(s) => s.clone(),
        _ => "NULL".to_owned(),
    }
}

async fn post_commands(
    http: impl AsRef<Http>,
    commands: Vec<Value>,
) -> serenity::Result<Vec<Command>> {
    let result = http
        .as_ref()
        .create_global_application_commands(&Value::from(commands))
        .await;

    match result {
        Ok(v) => {
            log::info!("Posted {} commands", v.len());
            log::debug!("Posted commands: {v:?}");
            Ok(v)
        }
        Err(e) => Err(e),
    }
}

async fn listen_loop(
    client: &Arc<SubClient>,
    http: Arc<Http>,
) -> Result<String, Box<dyn std::error::Error>> {
    let mut payloads = Vec::<Payload>::new();

    loop {
        tokio::select! {
            msg = client.recv() => {
                let msg = match msg {
                    Ok(v) => v,
                    Err(e) => {
                        break Err(e.into());
                    }
                };

                let payload: Value = match serde_json::from_str(&msg.payload) {
                    Ok(v) => v,
                    Err(e) => {
                        log::error!("Received poorly encoded pub/sub payload on channel 'commands': {e}");
                        continue;
                    }
                };
                let payload = match Payload::from_json(&payload) {
                    Ok(v) => v,
                    Err(e) => {
                        log::error!("Received invalid pub/sub payload on channel 'commands': {e}");
                        continue;
                    }
                };

                log::info!("Received command '{}'", display_value(payload.data.get("name")));
                log::debug!("Received command: {:?}", payload.data);

                payloads.push(payload);
            }
            _ = tokio::time::sleep(Duration::from_secs(20)) => {
                let s = match serde_json::to_string(&payloads) {
                    Ok(v) => v,
                    Err(e) => break Err(e.into()),
                };

                let commands = extract_payload_data(payloads);

                match post_commands(http, commands).await {
                    Ok(_) => {
                        break Ok(s);
                    }
                    Err(e) => {
                        break Err(e.into());
                    }
                };
            }
        }
    }
}

async fn entrypoint() -> Result<(), String> {
    let process_env = env::var("APP_ENV");

    if let Err(_) = process_env {
        dotenvy::dotenv().or(Err(
            "APP_ENV variable not provided and .env file could not be opened",
        ))?;
    }
    env_logger::init();

    let redis_uri = env::var("REDIS_URL").or(Err("REDIS_URL environment variable"))?;
    let token = env::var("DISCORD_TOKEN").or(Err("DISCORD_TOKEN environment variable"))?;
    let application_id: u64 = env::var("DISCORD_APP_ID")
        .or(Err("DISCORD_APP_ID environment variable"))?
        .parse()
        .or(Err(
            "DISCORD_APP_ID must be a valid 64 bit unsigned integer",
        ))?;

    let mut client = SubClient::connect(&redis_uri)
        .or_else(|e| Err(format!("Failed to connect to redis: {e}")))?;
    client.subscribe("commands");

    let client = Arc::new(client);
    spawn_daemon(client.clone());

    let http = Arc::new(Http::new(&token));
    http.set_application_id(application_id);

    let result = listen_loop(&client, http)
        .await
        .or_else(|e| Err(format!("{e}")))?;

    client
        .set("commands", result)
        .or_else(|e| Err(e.to_string()))?;

    Ok(())
}

#[tokio::main]
async fn main() {
    let result = entrypoint().await;

    match result {
        Ok(_) => {
            log::info!("Exited successfully");
            process::exit(0);
        }
        Err(e) => {
            log::error!("Failed to post application commands: {e}");
            process::exit(1);
        }
    }
}
