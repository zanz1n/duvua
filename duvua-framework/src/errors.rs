use crate::builder::interaction_response::InteractionResponse;
use serenity::{
    builder::{CreateComponents, CreateInteractionResponseData},
    model::prelude::{application_command::ApplicationCommandInteraction, InteractionResponseType},
    prelude::Context,
};

#[derive(thiserror::Error, Debug)]
pub enum BotError {
    #[error("Serenity error: {0}")]
    Serenity(serenity::Error),
    #[error("User could not be found")]
    UserNotFound,
    #[error("Something went wrong while performing a query")]
    Query,
    #[error("User already exists")]
    UserAlreadyExists,
}

impl BotError {
    pub fn get_message(&self) -> &'static str {
        ""
    }

    pub async fn respond_to(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
        defered: bool,
    ) {
        if let &BotError::Serenity(e) = &self {
            log::error!("Serenity error: {e}");
        } else {
            _ = InteractionResponse::default()
                .set_kind(if defered {
                    InteractionResponseType::UpdateMessage
                } else {
                    InteractionResponseType::ChannelMessageWithSource
                })
                .set_data(
                    CreateInteractionResponseData::default()
                        .set_components(CreateComponents::default())
                        .set_embeds(Vec::new())
                        .content(self.get_message()),
                )
                .respond_message_component(ctx.http.as_ref(), interaction)
                .await;
        }
    }
}
