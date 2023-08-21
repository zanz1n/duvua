use async_trait::async_trait;
use duvua_cache::{redis::RedisCacheService, CacheRepository};
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_option,
};
use serenity::{
    builder::{
        CreateApplicationCommand, CreateApplicationCommandOption, CreateButton, CreateEmbed,
        CreateInteractionResponseData,
    },
    model::{
        prelude::{
            application_command::ApplicationCommandInteraction, command::CommandOptionType,
            component::ButtonStyle, ReactionType,
        },
        user::User,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct AvatarCommand {
    data: &'static CommandHandlerData,
    cache: Arc<RedisCacheService>,
}

impl AvatarCommand {
    pub fn new(cache: Arc<RedisCacheService>) -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
            cache,
        }
    }
}

#[async_trait]
impl CommandHandler for AvatarCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let user_id = match get_option(&interaction.data.options, "user") {
            Some(v) => {
                let v = v
                    .value
                    .ok_or(BotError::InvalidOption("user"))?
                    .as_str()
                    .ok_or(BotError::InvalidOption("user"))?
                    .to_owned();

                v
            }
            None => interaction.user.id.to_string(),
        };

        let key = format!("user/{}", &user_id);

        let user: Option<User> = self.cache.de_get(key.clone()).await?;

        let user = match user {
            Some(v) => v,
            None => {
                let user = ctx
                    .http
                    .get_user(
                        user_id
                            .parse()
                            .or_else(|_| Err(BotError::InvalidOption("user")))?,
                    )
                    .await
                    .or_else(|e| Err(BotError::Serenity(e)))?;

                self.cache.ser_set_ttl(key, user.clone(), 60).await?;

                user
            }
        };

        let url = user.avatar_url().ok_or(BotError::UserAvatarFetchFailed)?;

        let mut response = InteractionResponse::default();
        response.set_data(
            CreateInteractionResponseData::default()
                .set_embed(
                    CreateEmbed::default()
                        .title(format!("Avatar de {}", user.name))
                        .image(&url)
                        .to_owned(),
                )
                .set_components(
                    CreateActionRow::default()
                        .add_button(
                            CreateButton::default()
                                .style(ButtonStyle::Link)
                                .label("Ver original")
                                .emoji(ReactionType::Unicode("🔗".to_owned()))
                                .url(&url)
                                .to_owned(),
                        )
                        .to_components(),
                )
                .to_owned(),
        );

        response
            .respond(&ctx.http, interaction.id.0, &interaction.token)
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

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("avatar")
        .description("Exibe o avatar de um usuário ou o seu")
        .description_localized("en-US", "Displays a user's avatar or yours")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::User)
                .name("user")
                .description("O usuário para ver o avatar")
                .description_localized("en-US", "Whose avatar you want to show")
                .to_owned(),
        )
        .to_owned()
}
