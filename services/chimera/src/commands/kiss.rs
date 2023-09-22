use async_trait::async_trait;
use duvua_cache::CacheRepository;
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
    utils::get_option,
};
use serenity::{
    builder::{
        CreateApplicationCommand, CreateApplicationCommandOption, CreateEmbed, CreateEmbedFooter,
        CreateInteractionResponseData,
    },
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};
use std::sync::Arc;

use crate::repository::{
    kiss_cache::KissCacheData,
    kiss_shared::{create_kiss_deny_button, create_kiss_reply_button},
    random::RandomStringProvider,
};

pub struct KissCommand<C: CacheRepository> {
    data: &'static CommandHandlerData,
    cache: Arc<C>,
    kiss_gifs: Arc<RandomStringProvider>,
}

impl<C: CacheRepository> KissCommand<C> {
    pub fn new(kiss_gifs: Arc<RandomStringProvider>, cache: Arc<C>) -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
            kiss_gifs,
            cache,
        }
    }
}

#[async_trait]
impl<C: CacheRepository> CommandHandler for KissCommand<C> {
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

        let rand_gif = self
            .kiss_gifs
            .get_choice()
            .ok_or(BotError::SomethingWentWrong)?;

        embed.image(rand_gif);

        if target_id == user_id {
            embed.set_footer(
                CreateEmbedFooter::default()
                    .text("Amar a si mesmo é bom!")
                    .to_owned(),
            );
        } else {
            let custom_id = nanoid::nanoid!(30);

            self.cache
                .ser_set_ttl(
                    "component/".to_owned() + &custom_id,
                    KissCacheData::new(
                        interaction.user.id.0,
                        target_id
                            .parse()
                            .or_else(|_| Err(BotError::SomethingWentWrong))?,
                        interaction.token.clone(),
                    ),
                    10,
                )
                .await?;

            response_data.set_components(
                CreateActionRow::default()
                    .add_button(create_kiss_reply_button(&custom_id, true))
                    .add_button(create_kiss_deny_button(&custom_id, true))
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
        command_data: Some(build_data_command()),
        custom_id: None,
        needs_defer: false,
        kind: CustomCommandType::Fun,
    }
}

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
