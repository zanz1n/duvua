use super::get_base64_image_data;
use async_trait::async_trait;
use duvua_cache::{redis::RedisCacheService, utils::get_or_store_user};
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
    utils::{get_avatar_url, get_option},
};
use serde_json::json;
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};
use std::{sync::Arc, time::Duration};

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
            .or(Err(BotError::InvalidOption("user")))?;

        let message = get_option(&interaction.data.options, "message")
            .ok_or(BotError::OptionNotProvided("message"))?
            .value
            .ok_or(BotError::InvalidOption("message"))?
            .as_str()
            .ok_or(BotError::InvalidOption("message"))?
            .to_owned();

        let user = get_or_store_user(&ctx.http, &self.cache, user_id).await?;

        let avatar = user.avatar.ok_or(BotError::UserAvatarFetchFailed)?;
        let avatar = get_avatar_url(user_id, &avatar, "webp", Some(128));

        let avatar_base64 = match get_base64_image_data(&avatar).await {
            Ok(v) => Some("data:image/webp;base64,".to_owned() + &v),
            Err(_) => None,
        };

        let data = json!({
            "name": user.name,
            "avatar": avatar_base64,
        });

        let webhook = ctx
            .http
            .create_webhook(interaction.channel_id.0, &data, Some("Clone command"))
            .await
            .or_else(|e| Err(BotError::Serenity(e)))?;

        _ = InteractionResponse::with_content_ephemeral("Clone criado")
            .respond(&ctx.http, interaction.id.0, &interaction.token)
            .await;

        if let Err(e) = webhook
            .execute(&ctx.http, false, |w| w.content(message))
            .await
        {
            log::error!("Failed to send webhook message: {e}");
        }

        tokio::time::sleep(Duration::from_millis(3000)).await;

        if let Err(e) = webhook.delete(&ctx.http).await {
            log::error!("Failed to delete webhook: {e}");
        }

        Ok(())
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        needs_defer: false,
        command_data: Some(build_data_command()),
        custom_id: None,
        kind: CustomCommandType::Fun,
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
