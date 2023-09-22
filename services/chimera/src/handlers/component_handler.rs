use crate::repository::{
    kiss_cache::KissCacheData,
    kiss_shared::{
        create_kiss_deny_button, create_kiss_reply_button, expiry_kiss_buttons, KissSharedHandler,
        SKIP_STR,
    },
};
use async_trait::async_trait;
use duvua_cache::CacheRepository;
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::CommandHandler,
};
use serenity::{
    builder::CreateInteractionResponseData,
    model::prelude::{message_component::MessageComponentInteraction, InteractionResponseType},
    prelude::Context,
};
use std::sync::Arc;

pub struct MessageComponentHandler<C: CacheRepository> {
    shared_handler: Arc<KissSharedHandler>,
    cache: Arc<C>,
}

impl<C: CacheRepository> MessageComponentHandler<C> {
    pub fn new(shared_handler: Arc<KissSharedHandler>, cache: Arc<C>) -> Self {
        Self {
            shared_handler,
            cache,
        }
    }

    async fn kiss_handler(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
        id_prefix: &str,
        id_rest: &str,
    ) -> Result<(), BotError> {
        if id_rest.len() != 30 {
            return Ok(());
        }

        let info: Option<KissCacheData> =
            self.cache.de_get("component/".to_owned() + id_rest).await?;
        let info = match info {
            Some(v) => v,
            None => {
                return InteractionResponse::default()
                    .set_kind(InteractionResponseType::UpdateMessage)
                    .set_data(
                        CreateInteractionResponseData::default()
                            .set_components(
                                CreateActionRow::default()
                                    .add_button(create_kiss_reply_button(SKIP_STR, false))
                                    .add_button(create_kiss_deny_button(SKIP_STR, false))
                                    .to_components(),
                            )
                            .to_owned(),
                    )
                    .respond_message_component(&ctx.http, interaction)
                    .await
            }
        };

        if info.target_id != interaction.user.id.0 {
            return InteractionResponse::with_content_ephemeral("Isso não é pra você enxerido!")
                .respond(&ctx.http, interaction.id.0, &interaction.token)
                .await;
        }

        let reponse = if id_prefix == "rkiss" {
            self.shared_handler
                .handle_kiss_reply(info.user_id, info.target_id)
                .await?
        } else {
            self.shared_handler
                .handle_kiss_deny(info.user_id, info.target_id)
                .await?
        };

        self.cache.del("component/".to_owned() + id_rest).await?;
        expiry_kiss_buttons(&ctx.http, &info.interaction_token).await?;

        reponse
            .respond(&ctx.http, interaction.id.0, &interaction.token)
            .await
    }
}

#[async_trait]
impl<C: CacheRepository> CommandHandler for MessageComponentHandler<C> {
    async fn handle_component(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        let custom_id = interaction.data.custom_id.as_str();

        if custom_id.len() <= 33 {
            return Ok(());
        }

        let prefix = &custom_id[0..5];

        if prefix == "rkiss" || prefix == "dkiss" {
            self.kiss_handler(ctx, interaction, prefix, &custom_id[6..])
                .await
        } else {
            Ok(())
        }
    }
}
