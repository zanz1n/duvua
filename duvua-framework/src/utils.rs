use crate::builder::hashmap_to_json_map;
use serde_json::Value;
use serenity::{
    builder::CreateMessage,
    http::Http,
    model::prelude::{application_command::CommandDataOption, command::CommandOptionType, Message},
};

#[inline]
pub fn get_sub_command(options: &Vec<CommandDataOption>) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.kind == CommandOptionType::SubCommand {
            return Some(option.clone());
        }
    }

    None
}

#[inline]
pub fn get_sub_command_group(options: &Vec<CommandDataOption>) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.kind == CommandOptionType::SubCommandGroup {
            return Some(option.clone());
        }
    }

    None
}

#[inline]
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
