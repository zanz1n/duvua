use super::random::RandomStringProvider;
use duvua_framework::{
    builder::interaction_response::InteractionResponse, errors::BotError,
    utils::update_interaction_response,
};
use serenity::{
    builder::{
        CreateActionRow, CreateButton, CreateEmbed, CreateInteractionResponseData,
        EditInteractionResponse,
    },
    http::Http,
    model::prelude::{component::ButtonStyle, ReactionType},
};
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
                            .description(format!(
                                "<@{target_id}> negou o beijo de <@{user_id}>  💔",
                            ))
                            .image(image_url)
                            .to_owned(),
                    )
                    .to_owned(),
            )
            .to_owned())
    }
}

pub fn create_kiss_reply_button(custom_id: &str, enabled: bool) -> CreateButton {
    let mut button = CreateButton::default();

    button
        .style(ButtonStyle::Primary)
        .label("Retribuir")
        .emoji(ReactionType::Unicode("🔁".to_owned()))
        .custom_id("rkiss/".to_owned() + custom_id)
        .disabled(!enabled);
    button
}

pub fn create_kiss_deny_button(custom_id: &str, enabled: bool) -> CreateButton {
    let mut button = CreateButton::default();

    button
        .style(ButtonStyle::Primary)
        .label("Recusar")
        .emoji(ReactionType::Unicode("❌".to_owned()))
        .custom_id("dkiss/".to_owned() + custom_id)
        .disabled(!enabled);
    button
}

pub const SKIP_STR: &str = "9DZ7pNKqYL4j2vHJz4Hb3oCITFGN1vGNlNjIuPUZb0+xtaSTyT+32Ew1atBTfANH";

pub async fn expiry_kiss_buttons(
    http: impl AsRef<Http>,
    interaction_token: &str,
) -> Result<(), BotError> {
    let mut data = EditInteractionResponse::default();

    data.components(|c| {
        let mut action_row = CreateActionRow::default();

        action_row
            .add_button(create_kiss_reply_button(SKIP_STR, false))
            .add_button(create_kiss_deny_button(SKIP_STR, false));

        c.set_action_row(action_row)
    });

    match update_interaction_response(http, interaction_token, data).await {
        Ok(_) => Ok(()),
        Err(e) => Err(BotError::Serenity(e)),
    }
}
