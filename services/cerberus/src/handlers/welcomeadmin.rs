use super::welcome_shared::WelcomeSharedHandler;
use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
    utils::{get_option, get_sub_command},
};
use duvua_repository::welcome::WelcomeType;
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    http::Http,
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType, Member,
    },
    prelude::Context,
};
use std::{str::FromStr, sync::Arc};

pub struct WelcomeAdminCommand {
    shared_handler: Arc<WelcomeSharedHandler>,
    data: &'static CommandHandlerData,
}

impl WelcomeAdminCommand {
    pub fn new(shared_handler: Arc<WelcomeSharedHandler>) -> Self {
        Self {
            shared_handler,
            data: Box::leak(Box::new(build_data())),
        }
    }

    pub async fn handle_test_message(
        &self,
        http: &Arc<Http>,
        guild_id: u64,
        member: &Member,
    ) -> Result<InteractionResponse, BotError> {
        let bool_result = self.shared_handler.trigger(http, guild_id, member).await?;

        let msg = if bool_result {
            "Mensagem de teste enviada!"
        } else {
            "É necessário definir um canal de texto para enviar as mensagens. Use `/welcomeadmin set-channel`."
        }
        .to_owned();

        Ok(InteractionResponse::with_content(msg))
    }
}

#[async_trait]
impl CommandHandler for WelcomeAdminCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let guild_id = interaction
            .guild_id
            .ok_or(BotError::CommandIssuedOutOfGuild)?
            .0;

        let member = interaction
            .member
            .as_ref()
            .ok_or(BotError::CommandIssuedOutOfGuild)?;

        let permissions = member
            .permissions
            .ok_or(BotError::CommandIssuedByPartialMember)?;

        if !permissions.administrator() {
            return Err(BotError::CommandPermissionDenied);
        }

        let sub_command = get_sub_command(&interaction.data.options)
            .ok_or(BotError::OptionNotProvided("sub-command"))?;

        let response = match sub_command.name.as_str() {
            "enable" => {
                self.shared_handler
                    .handle_set_enabled(guild_id, true)
                    .await?
            }
            "disable" => {
                self.shared_handler
                    .handle_set_enabled(guild_id, false)
                    .await?
            }
            "set-channel" => {
                let channel_id = get_option(&sub_command.options, "channel")
                    .ok_or(BotError::OptionNotProvided("channel"))?
                    .value
                    .ok_or(BotError::InvalidOption("channel"))?
                    .as_str()
                    .ok_or(BotError::InvalidOption("channel"))?
                    .parse()
                    .or(Err(BotError::InvalidOption("channel")))?;

                self.shared_handler
                    .handle_set_channel(&ctx.http, guild_id, channel_id)
                    .await?
            }
            "set-message" => {
                let message = get_option(&sub_command.options, "message")
                    .ok_or(BotError::OptionNotProvided("message"))?
                    .value
                    .ok_or(BotError::OptionNotProvided("message"))?
                    .as_str()
                    .ok_or(BotError::OptionNotProvided("message"))?
                    .to_owned();

                let kind = get_option(&sub_command.options, "type")
                    .ok_or(BotError::OptionNotProvided("type"))?
                    .value
                    .ok_or(BotError::InvalidOption("type"))?
                    .as_str()
                    .ok_or(BotError::InvalidOption("type"))?
                    .to_owned();

                let kind = WelcomeType::from_str(&kind).or(Err(BotError::InvalidOption("type")))?;

                self.shared_handler
                    .handle_set_message(guild_id, kind, message)
                    .await?
            }
            "test" => {
                self.handle_test_message(&ctx.http, guild_id, member)
                    .await?
            }
            _ => return Err(BotError::InvalidOption("sub-command")),
        };

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
        needs_defer: false,
        command_data: Some(build_data_command()),
        custom_id: None,
        kind: CustomCommandType::Config,
    }
}

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("welcomeadmin")
        .description("Comandos para configurar a funcionalidade de boas vindas")
        .description_localized("en-US", "Commands to configure welcome functionality")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("disable")
                .description("Disabilita mensagens de boas vindas no servidor")
                .description_localized("en-US", "Disables the welcome message on the server")
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("enable")
                .description("Habilita mensagens de boas vindas no servidor")
                .description_localized("en-US", "Enables the welcome message on the server")
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("set-channel")
                .description(
                    "Atualiza o canal de texto no qual as mensagens de boas vindas são enviadas",
                )
                .description_localized("en-US", "Updates the welcome messages text channel")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::Channel)
                        .name("channel")
                        .description("O canal de texto")
                        .description_localized("en-US", "The text channel")
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("set-message")
                .description("Atualiza a mensagem de boas vindas que é enviada")
                .description_localized("en-US", "Updates the welcome message that will be sent")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::String)
                        .name("type")
                        .description("O tipo de mensagem que é enviada")
                        .description_localized("en-US", "The type of the message that will be sent")
                        .add_string_choice("Message", "MESSAGE")
                        .add_string_choice("Image", "IMAGE")
                        .add_string_choice("Embed", "EMBED")
                        .required(true)
                        .to_owned(),
                )
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::String)
                        .name("message")
                        .description(
                            "Placeholders: {{USER}} (o novo membro), \
                            {{GUILD}} (nome do servidor) podem ser usados",
                        )
                        .description_localized(
                            "en-US",
                            "Placeholders: {{USER}} (the new member) \
                            {{GUILD}} (the name of server) can be used",
                        )
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("test")
                .description("Testa a mensagem de boas vindas do servidor")
                .description_localized("en-US", "Tests the welcome messsage")
                .to_owned(),
        )
        .to_owned()
}
