use crate::builder::hashmap_to_json_map;
use serde_json::Value;
use serenity::{
    builder::{CreateMessage, EditInteractionResponse},
    http::Http,
    model::prelude::{application_command::CommandDataOption, command::CommandOptionType, Message},
};

pub fn get_avatar_url(user_id: u64, hash: &str, ext: &str, size: Option<u16>) -> String {
    let mut query = String::new();

    if let Some(size) = size {
        query += "?size=";
        query += &size.to_string()
    }

    format!("https://cdn.discordapp.com/avatars/{user_id}/{hash}.{ext}{query}")
}

#[inline]
pub fn get_cdn_address(addr: &str) -> String {
    "https://cdn.discordapp.com".to_owned() + addr
}

pub fn get_sub_command(options: &Vec<CommandDataOption>) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.kind == CommandOptionType::SubCommand {
            return Some(option.clone());
        }
    }

    None
}

pub fn get_sub_command_group(options: &Vec<CommandDataOption>) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.kind == CommandOptionType::SubCommandGroup {
            return Some(option.clone());
        }
    }

    None
}

pub fn get_option<T: ToString>(
    options: &Vec<CommandDataOption>,
    name: T,
) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.name == name.to_string() {
            return Some(option.clone());
        }
    }

    None
}

pub async fn send_message(
    http: impl AsRef<Http>,
    msg: CreateMessage<'_>,
    channel_id: u64,
) -> serenity::Result<Message> {
    let map = hashmap_to_json_map(msg.0);

    let message = if msg.2.is_empty() {
        http.as_ref()
            .send_message(channel_id, &Value::from(map))
            .await?
    } else {
        http.as_ref().send_files(channel_id, msg.2, &map).await?
    };

    if let Some(reactions) = msg.1 {
        for reaction in reactions {
            http.as_ref()
                .create_reaction(channel_id, message.id.0, &reaction)
                .await?;
        }
    }

    Ok(message)
}

pub async fn update_interaction_response(
    http: impl AsRef<Http>,
    interaction_token: &str,
    data: EditInteractionResponse,
) -> serenity::Result<Message> {
    let map = hashmap_to_json_map(data.0);
    http.as_ref()
        .edit_original_interaction_response(interaction_token, &Value::Object(map))
        .await
}
