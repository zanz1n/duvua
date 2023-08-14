use crate::repository::kiss_shared::KissSharedHandler;
use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse, errors::BotError, handler::CommandHandler,
};
use serenity::{model::prelude::message_component::MessageComponentInteraction, prelude::Context};
use std::sync::Arc;

pub struct MessageComponentHandler {
    shared_handler: Arc<KissSharedHandler>,
}

impl MessageComponentHandler {
    pub fn new(shared_handler: Arc<KissSharedHandler>) -> Self {
        Self { shared_handler }
    }
}

#[async_trait]
impl CommandHandler for MessageComponentHandler {
    async fn handle_component(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        let custom_id = interaction.data.custom_id.as_str();

        if custom_id.len() <= 38 {
            return Ok(());
        }

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

            if target_id != interaction.user.id.0 {
                return InteractionResponse::with_content_ephemeral(
                    "Isso não é pra você enxerido!",
                )
                .respond(&ctx.http, interaction.id.0, &interaction.token)
                .await;
            }

            let reponse = if prefix == "rkiss" {
                self.shared_handler
                    .handle_kiss_reply(user_id, target_id)
                    .await?
            } else {
                self.shared_handler
                    .handle_kiss_deny(user_id, target_id)
                    .await?
            };

            reponse
                .respond(&ctx.http, interaction.id.0, &interaction.token)
                .await
        } else {
            Ok(())
        }
    }
}
