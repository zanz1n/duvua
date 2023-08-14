use super::random::RandomStringProvider;
use duvua_framework::{builder::interaction_response::InteractionResponse, errors::BotError};
use serenity::builder::{CreateEmbed, CreateInteractionResponseData};
use std::sync::Arc;

pub struct KissSharedHandler {
    kiss_gifs: Arc<RandomStringProvider>,
    slap_gifs: Arc<RandomStringProvider>,
}

impl KissSharedHandler {
    pub fn new(kiss_gifs: Arc<RandomStringProvider>, slap_gifs: Arc<RandomStringProvider>) -> Self {
        Self {
            kiss_gifs,
            slap_gifs,
        }
    }

    pub async fn handle_kiss_reply(
        &self,
        user_id: u64,
        target_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let image_url = self
            .kiss_gifs
            .get_choice()
            .ok_or(BotError::SomethingWentWrong)?;

        Ok(InteractionResponse::default()
            .set_data(
                CreateInteractionResponseData::default()
                    .set_embed(
                        CreateEmbed::default()
                            .title("As coisas estão pegando fogo!  🔥")
                            .description(format!(
                                "<@{target_id}> retribuiu o beijo de <@{user_id}>\n\
                                Será que temos um novo casal aqui?  ❤️",
                            ))
                            .image(image_url)
                            .to_owned(),
                    )
                    .to_owned(),
            )
            .to_owned())
    }

    pub async fn handle_kiss_deny(
        &self,
        user_id: u64,
        target_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let image_url = self
            .slap_gifs
            .get_choice()
            .ok_or(BotError::SomethingWentWrong)?;

        Ok(InteractionResponse::default()
            .set_data(
                CreateInteractionResponseData::default()
                    .set_embed(
                        CreateEmbed::default()
                            .title("Quem nunca levou um fora, né?")
                            .description(format!("<@{target_id}> negou o beijo de {user_id}  💔",))
                            .image(image_url)
                            .to_owned(),
                    )
                    .to_owned(),
            )
            .to_owned())
    }
}
