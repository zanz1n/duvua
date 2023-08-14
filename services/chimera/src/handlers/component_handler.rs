use async_trait::async_trait;
use duvua_framework::{errors::BotError, handler::CommandHandler};
use serenity::{model::prelude::message_component::MessageComponentInteraction, prelude::Context};

pub struct MessageComponentHandler {}

impl MessageComponentHandler {
    pub fn new() -> Self {
        Self {}
    }
}

#[async_trait]
impl CommandHandler for MessageComponentHandler {
    async fn handle_component(
        &self,
        _ctx: &Context,
        interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        let custom_id = interaction.data.custom_id.as_str();

        if custom_id.len() > 38 {
            let prefix = &custom_id[0..5];

            if prefix == "rkiss" || prefix == "dkiss" {
                let rest = &custom_id[6..];

                let (user_id, target_id) = match rest.split_once('-') {
                    Some(v) => v,
                    None => return Ok(()),
                };

                let user_id: u64 = match user_id.parse() {
                    Ok(v) => v,
                    Err(_) => return Ok(()),
                };
                let target_id: u64 = match target_id.parse() {
                    Ok(v) => v,
                    Err(_) => return Ok(()),
                };

                _ = user_id;
                _ = target_id;
            }
        }

        Ok(())
    }
}
