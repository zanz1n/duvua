use super::ticketadmin_shared::TicketAdminSharedHandler;
use async_trait::async_trait;
use duvua_framework::{
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_sub_command,
};
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct TicketAdminCommandHandler {
    shared_handler: Arc<TicketAdminSharedHandler>,
    data: &'static CommandHandlerData,
}

impl TicketAdminCommandHandler {
    pub fn new(shared_handler: Arc<TicketAdminSharedHandler>) -> Self {
        Self {
            shared_handler,
            data: Box::leak(Box::new(build_data())),
        }
    }
}

#[async_trait]
impl CommandHandler for TicketAdminCommandHandler {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let member_permissions = interaction
            .member
            .as_ref()
            .ok_or(BotError::CommandIssuedOutOfGuild)?
            .permissions
            .ok_or(BotError::CommandPermissionDenied)?;

        let guild_id = interaction
            .guild_id
            .ok_or(BotError::CommandIssuedOutOfGuild)?
            .0;

        if !member_permissions.administrator() {
            return Err(BotError::CommandPermissionDenied);
        }

        let sub_command =
            get_sub_command(&interaction.data.options).ok_or(BotError::SomethingWentWrong)?;

        let res = match sub_command.name.as_str() {
            "disable" => {
                self.shared_handler
                    .handle_toggle_guild_tickets(guild_id, false)
                    .await?
            }
            "enable" => {
                self.shared_handler
                    .handle_toggle_guild_tickets(guild_id, true)
                    .await?
            }
            "delete" => {
                self.shared_handler
                    .handle_delete_ticket_by_options(&ctx.http, &sub_command.options)
                    .await?
            }
            "delete-all" => {
                self.shared_handler
                    .handle_delete_all(&ctx.http, guild_id, interaction.user.id.0)
                    .await?
            }
            "add-permanent" => {
                self.shared_handler
                    .handle_add_permanent_by_options(
                        &ctx.http,
                        interaction.user.id.0,
                        &sub_command.options,
                    )
                    .await?
            }
            _ => return Err(BotError::InvalidOption("sub-command")),
        };

        res.respond(&ctx.http, interaction.id.0, &interaction.token)
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
        .name("ticketadmin")
        .description("Comandos administrativos de tickets")
        .description_localized("en-US", "Ticket administrative commands")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("disable")
                .description("Desabilita a funcionalidade de tickets no servidor")
                .description_localized("en-US", "Disables the ticket functionality on the server")
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("enable")
                .description("Habilita a funcionalidade de tickets no servidor")
                .description_localized("en-US", "Enables the ticket functionality on the server")
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("add-permanent")
                .description("Posta uma mensagem com um botão para criar tickets no servidor")
                .description_localized(
                    "en-US",
                    "Posts a message with a button to create tickets in the server",
                )
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::Channel)
                        .name("channel")
                        .description("O canal para enviar a mensagem")
                        .description_localized("en-US", "The channel the message will be sent")
                        .required(true)
                        .to_owned(),
                )
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::String)
                        .name("message")
                        .description("A mensagem que será enviada")
                        .description_localized("en-US", "The message")
                        .to_owned(),
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("delete")
                .description("Deleta o ticket de um membro por id")
                .description_localized("en-US", "Deletes one member's ticket by id")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::String)
                        .name("id")
                        .description("O id do ticket que deseja deletar")
                        .description_localized("en-US", "The id of the ticket you want to delete")
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("delete-all")
                .description("Deleta todos os seus tickets de um membro")
                .description_localized("en-US", "Delete all the tickets of a member")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::User)
                        .name("member")
                        .description("O membro")
                        .description_localized("en-US", "The member")
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .to_owned()
}
