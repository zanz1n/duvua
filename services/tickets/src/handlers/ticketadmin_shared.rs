use crate::repository::{
    guild::{GuildRepository, UpdateGuildData},
    ticket::TicketRepository,
};
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    utils::{get_option, send_message},
};
use serenity::{
    builder::{CreateButton, CreateEmbed, CreateInteractionResponseData, CreateMessage},
    http::Http,
    model::prelude::{
        application_command::CommandDataOption, component::ButtonStyle, ReactionType,
    },
};
use std::sync::Arc;

pub struct TicketAdminSharedHandler {
    ticket_repo: Arc<dyn TicketRepository>,
    guild_repo: Arc<dyn GuildRepository>,
}

impl TicketAdminSharedHandler {
    pub fn new(
        ticket_repo: Arc<dyn TicketRepository>,
        guild_repo: Arc<dyn GuildRepository>,
    ) -> Self {
        Self {
            ticket_repo,
            guild_repo,
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

        Ok(InteractionResponse::default()
            .set_data(CreateInteractionResponseData::default().content(msg))
            .to_owned())
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

        Ok(InteractionResponse::default()
            .set_data(CreateInteractionResponseData::default().content(format!(
                "Ticket de <@{}> deletado com sucesso",
                ticket.user_id
            )))
            .to_owned())
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

        Ok(InteractionResponse::default()
            .set_data(CreateInteractionResponseData::default().content("Mensagem enviada"))
            .to_owned())
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

        log::debug!("{channel:?}");

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
}
