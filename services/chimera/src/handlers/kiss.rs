use crate::repository::KISS_GIFS;
use async_trait::async_trait;
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_option,
};
use rand::seq::SliceRandom;
use serenity::{
    builder::{
        CreateApplicationCommand, CreateApplicationCommandOption, CreateButton, CreateEmbed,
        CreateEmbedFooter, CreateInteractionResponseData,
    },
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
        component::ButtonStyle, ReactionType,
    },
    prelude::Context,
};

pub struct KissCommand {
    data: &'static CommandHandlerData,
    kiss_gifs: Vec<&'static str>,
}

impl KissCommand {
    pub fn new() -> Self {
        let kiss_gifs: Vec<&str> = KISS_GIFS.split("\n").collect();
        Self {
            data: Box::leak(Box::new(build_data())),
            kiss_gifs,
        }
    }
}

#[async_trait]
impl CommandHandler for KissCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let user_id = interaction.user.id.0.to_string();

        let target_id = match get_option(&interaction.data.options, "user") {
            Some(option) => option
                .value
                .ok_or(BotError::InvalidOption("user"))?
                .as_str()
                .ok_or(BotError::InvalidOption("user"))?
                .to_owned(),
            None => user_id.clone(),
        };

        let mut response_data = CreateInteractionResponseData::default();

        let mut embed = CreateEmbed::default();
        embed
            .title("O amor está no ar!  ❤️")
            .description(format!("<@{user_id}> beijou <@{target_id}>"));

        let rand_gif = *self
            .kiss_gifs
            .choose(&mut rand::thread_rng())
            .ok_or(BotError::SomethingWentWrong)?;

        embed.image(rand_gif);

        if target_id == user_id {
            embed.set_footer(
                CreateEmbedFooter::default()
                    .text("Amar a si mesmo é bom!")
                    .to_owned(),
            );
        } else {
            response_data.set_components(
                CreateActionRow::default()
                    .add_button(
                        CreateButton::default()
                            .style(ButtonStyle::Primary)
                            .label("Retribuir")
                            .emoji(ReactionType::Unicode("🔁".to_owned()))
                            .custom_id(format!("rkiss/{user_id}-{target_id}"))
                            .to_owned(),
                    )
                    .add_button(
                        CreateButton::default()
                            .style(ButtonStyle::Primary)
                            .label("Recusar")
                            .emoji(ReactionType::Unicode("❌".to_owned()))
                            .custom_id(format!("dkiss/{user_id}-{target_id}"))
                            .to_owned(),
                    )
                    .to_components(),
            );
        }

        response_data.set_embed(embed);

        InteractionResponse::default()
            .set_data(response_data)
            .respond_application_command(&ctx.http, interaction)
            .await
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        accepts_application_command: true,
        accepts_message_component: false,
        command_data: Some(build_data_command()),
        custom_id: None,
        needs_defer: false,
    }
}

#[inline]
fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("kiss")
        .description("demonstre seu amor com outro membro do servidor beijando-o")
        .description_localized(
            "en-US",
            "show your love to another server member by kissing him",
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::User)
                .name("user")
                .description("O usuário que deseija beijar")
                .description_localized("en-US", "The user you want to kiss")
                .required(true)
                .to_owned(),
        )
        .to_owned()
}
