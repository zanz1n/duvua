use crate::repository::{
    guild::{GuildRepository, GuildService, UpdateGuildData},
    ticket::{TicketRepository, TicketService},
};
use async_trait::async_trait;
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
    utils::{get_option, get_sub_command, send_message},
};
use serenity::{
    builder::{
        CreateApplicationCommand, CreateApplicationCommandOption, CreateButton, CreateEmbed,
        CreateMessage,
    },
    http::Http,
    model::prelude::{
        application_command::{ApplicationCommandInteraction, CommandDataOption},
        command::CommandOptionType,
        component::ButtonStyle,
        ReactionType,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct TicketAdminCommandHandler {
    ticket_repo: Arc<TicketService>,
    guild_repo: Arc<GuildService>,
    data: &'static CommandHandlerData,
}

impl TicketAdminCommandHandler {
    pub fn new(ticket_repo: Arc<TicketService>, guild_repo: Arc<GuildService>) -> Self {
        Self {
            guild_repo,
            ticket_repo,
            data: Box::leak(Box::new(build_data())),
        }
    }

    pub async fn handle_toggle_guild_tickets(
        &self,
        guild_id: u64,
        enabled: bool,
    ) -> Result<InteractionResponse, BotError> {
        self.guild_repo
            .update(
                guild_id,
                UpdateGuildData {
                    allow_multiple: None,
                    channel_category: None,
                    enable_tickets: Some(enabled),
                },
            )
            .await?;

        let msg = if enabled {
            "Tickets habilidatos no servidor"
        } else {
            "Ticket desabilidatos no servidor"
        };

        Ok(InteractionResponse::with_content(msg))
    }

    pub async fn handle_delete_ticket_by_id(
        &self,
        http: impl AsRef<Http>,
        id: String,
    ) -> Result<InteractionResponse, BotError> {
        let ticket = self.ticket_repo.get(id.clone()).await?;

        match http.as_ref().delete_channel(ticket.channel_id as u64).await {
            Ok(c) => log::info!("Channel {} deleted", c.id().0),
            Err(e) => log::warn!("Failed to delete channel: {e}"),
        }

        self.ticket_repo.delete(id).await?;

        Ok(InteractionResponse::with_content(format!(
            "Ticket de <@{}> deletado com sucesso",
            ticket.user_id
        )))
    }

    pub async fn handle_delete_ticket_by_options(
        &self,
        http: impl AsRef<Http>,
        options: &Vec<CommandDataOption>,
    ) -> Result<InteractionResponse, BotError> {
        let id = get_option(options, "id")
            .ok_or(BotError::OptionNotProvided("id"))?
            .value
            .ok_or(BotError::InvalidOption("id"))?
            .as_str()
            .ok_or(BotError::InvalidOption("id"))?
            .to_owned();

        self.handle_delete_ticket_by_id(http, id).await
    }

    pub async fn handle_add_permanent(
        &self,
        http: impl AsRef<Http>,
        user_id: u64,
        message: Option<String>,
        channel_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let message = match message {
            Some(message) => format!("{message}\n- <@{user_id}>"),
            None => "Clique no botão para criar um ticket!".to_owned(),
        };

        let msg = CreateMessage::default()
            .set_embed(
                CreateEmbed::default()
                    .title("Tickets")
                    .description(message)
                    .to_owned(),
            )
            .set_components(
                CreateActionRow::default()
                    .add_button(
                        CreateButton::default()
                            .style(ButtonStyle::Primary)
                            .label("Criar Ticket")
                            .emoji(ReactionType::Unicode("🎫".to_owned()))
                            .custom_id("ticket")
                            .to_owned(),
                    )
                    .to_components(),
            )
            .to_owned();

        send_message(http, msg, channel_id)
            .await
            .or_else(|_| Err(BotError::FailedToSendChannelMessage))?;

        Ok(InteractionResponse::with_content("Mensagem enviada"))
    }

    pub async fn handle_add_permanent_by_options(
        &self,
        http: impl AsRef<Http>,
        user_id: u64,
        options: &Vec<CommandDataOption>,
    ) -> Result<InteractionResponse, BotError> {
        let channel = get_option(options, "channel")
            .ok_or(BotError::OptionNotProvided("channel"))?
            .value
            .ok_or(BotError::InvalidOption("channel"))?;

        let channel = channel
            .as_str()
            .ok_or(BotError::InvalidOption("channel"))?
            .parse()
            .or_else(|_| Err(BotError::InvalidOption("channel")))?;

        let message = match get_option(options, "message") {
            Some(message) => {
                let value = message.value.ok_or(BotError::InvalidOption("message"))?;
                let value = value.as_str().ok_or(BotError::InvalidOption("message"))?;

                Some(value.to_owned())
            }
            None => None,
        };

        self.handle_add_permanent(http, user_id, message, channel)
            .await
    }

    pub async fn handle_delete_all(
        &self,
        http: &Arc<Http>,
        guild_id: u64,
        user_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let tickets = self
            .ticket_repo
            .get_by_member(guild_id, user_id, 11)
            .await?;

        if tickets.len() > 10 {
            return Ok(InteractionResponse::with_content(
                "O usuário tem mais de 10 tickets, exclua eles manualmente",
            ));
        }

        let deleted_count = self.ticket_repo.delete_by_member(guild_id, user_id).await?;

        for ticket in tickets {
            let http = http.clone();

            tokio::spawn(async move {
                match http.delete_channel(ticket.channel_id as u64).await {
                    Ok(c) => log::info!("Channel {} on guild {guild_id} deleted", c.id().0),
                    Err(e) => log::warn!("Failed to delete channel: {e}"),
                }
            });
        }

        Ok(InteractionResponse::with_content(format!(
            "{deleted_count} tickets excluídos com sucesso"
        )))
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
            "disable" => self.handle_toggle_guild_tickets(guild_id, false).await?,
            "enable" => self.handle_toggle_guild_tickets(guild_id, true).await?,
            "delete" => {
                self.handle_delete_ticket_by_options(&ctx.http, &sub_command.options)
                    .await?
            }
            "delete-all" => {
                self.handle_delete_all(&ctx.http, guild_id, interaction.user.id.0)
                    .await?
            }
            "add-permanent" => {
                self.handle_add_permanent_by_options(
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
        command_data: Some(build_data_command()),
        custom_id: None,
        needs_defer: false,
        kind: CustomCommandType::Config,
    }
}

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
