use super::KISS_GIFS;
use duvua_framework::{builder::interaction_response::InteractionResponse, errors::BotError};
use rand::seq::SliceRandom;
use serenity::builder::{CreateEmbed, CreateInteractionResponseData};

pub struct KissSharedHandler {
    kiss_gifs: Vec<&'static str>,
}

impl KissSharedHandler {
    pub fn new() -> Self {
        let kiss_gifs: Vec<&str> = KISS_GIFS.split("\n").collect();
        Self { kiss_gifs }
    }

    pub async fn handle_kiss_reply(
        &self,
        user_id: u64,
        target_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let image_url = *self
            .kiss_gifs
            .choose(&mut rand::thread_rng())
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
        todo!()
    }
}
