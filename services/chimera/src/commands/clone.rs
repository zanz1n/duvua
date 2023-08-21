use async_trait::async_trait;
use duvua_cache::{redis::RedisCacheService, utils::get_or_store_user};
use duvua_framework::{
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_option,
};
use serde_json::json;
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct CloneCommand {
    data: &'static CommandHandlerData,
    cache: Arc<RedisCacheService>,
}

impl CloneCommand {
    pub fn new(cache: Arc<RedisCacheService>) -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
            cache,
        }
    }
}

#[async_trait]
impl CommandHandler for CloneCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let user_id: u64 = get_option(&interaction.data.options, "user")
            .ok_or(BotError::OptionNotProvided("user"))?
            .value
            .ok_or(BotError::InvalidOption("user"))?
            .as_str()
            .ok_or(BotError::InvalidOption("user"))?
            .parse()
            .or_else(|_| Err(BotError::InvalidOption("user")))?;

        let message = get_option(&interaction.data.options, "message")
            .ok_or(BotError::OptionNotProvided("message"))?
            .value
            .ok_or(BotError::InvalidOption("message"))?
            .as_str()
            .ok_or(BotError::InvalidOption("message"))?
            .to_owned();

        let user = get_or_store_user(&ctx.http, &self.cache, user_id).await?;

        let data = json!({
            "name": user.name,
            "avatar": user.static_avatar_url(),
        });

        let webhook = ctx
            .http
            .create_webhook(interaction.channel_id.0, &data, Some("Clone command"))
            .await
            .or_else(|e| Err(BotError::Serenity(e)))?;

        webhook
            .execute(&ctx.http, false, |w| w.content(message))
            .await
            .or_else(|e| Err(BotError::Serenity(e)))?;

        Ok(())
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        accepts_message_component: false,
        accepts_application_command: true,
        needs_defer: false,
        command_data: Some(build_data_command()),
        custom_id: None,
    }
}

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("clone")
        .description("Clone um usuário e faça o clone enviar uma mensagem")
        .description_localized("en-US", "Create a user clone and make it send some message")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::User)
                .name("user")
                .description("O usuário que deseja clonar")
                .description_localized("en-US", "The user you want to clone")
                .required(true)
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::String)
                .name("message")
                .description("A mensagem que o clone irá mandar")
                .description_localized("en-US", "The message's the clone will send")
                .required(true)
                .to_owned(),
        )
        .to_owned()
}
