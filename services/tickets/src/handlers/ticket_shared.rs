use crate::repository::{
    guild::Guild,
    ticket::{CreateTicketData, TicketRepository},
};
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    utils::{get_option, send_message},
};
use serenity::{
    builder::{
        CreateButton, CreateChannel, CreateEmbed, CreateInteractionResponseData, CreateMessage,
    },
    http::Http,
    json::hashmap_to_json_map,
    model::{
        prelude::{
            application_command::CommandDataOption, component::ButtonStyle, ChannelType,
            InteractionResponseType, PermissionOverwrite, PermissionOverwriteType, ReactionType,
            RoleId, UserId,
        },
        Permissions,
    },
};
use std::sync::Arc;

pub struct TicketSharedHandler {
    ticket_repo: Arc<dyn TicketRepository>,
}

impl TicketSharedHandler {
    pub fn new(ticket_repo: Arc<dyn TicketRepository>) -> Self {
        Self { ticket_repo }
    }

    pub async fn handle_create_ticket(
        &self,
        http: impl AsRef<Http>,
        guild: Guild,
        guild_id: u64,
        user_id: u64,
        username: String,
    ) -> Result<InteractionResponse, BotError> {
        if !guild.allow_multiple {
            let ticket = self.ticket_repo.get_by_member(guild_id, user_id, 2).await?;
            if ticket.len() != 0 {
                return Err(BotError::OnlyOneTicketAllowed);
            }
        }

        let mut data = CreateTicketData::from_snowflakes(0, user_id, guild_id);
        let hex_id = data.id.to_hex();

        let permissions = [
            PermissionOverwrite {
                allow: Permissions::SEND_MESSAGES.union(Permissions::VIEW_CHANNEL),
                deny: Permissions::default(),
                kind: PermissionOverwriteType::Member(UserId::from(user_id)),
            },
            PermissionOverwrite {
                allow: Permissions::default(),
                deny: Permissions::SEND_MESSAGES.union(Permissions::VIEW_CHANNEL),
                kind: PermissionOverwriteType::Role(RoleId(guild_id)),
            },
        ];

        let mut channel = CreateChannel::default()
            .kind(ChannelType::Text)
            .name(username + "-" + data.id.to_hex().as_str())
            .permissions(permissions)
            .to_owned();

        if let Some(cat) = guild.channel_category {
            channel.category(cat);
        }

        let channel = http
            .as_ref()
            .create_channel(
                guild_id,
                &hashmap_to_json_map(channel.0),
                Some(&format!("Ticket {hex_id}")),
            )
            .await
            .or_else(|e| Err(BotError::Serenity(e)))?;

        data.channel_id = channel.id.0 as i64;
        let id = data.id.to_hex();

        let msg = CreateMessage::default()
            .set_embed(
                CreateEmbed::default()
                    .title("Ticket criado")
                    .description(format!(
                        "ID: `{id}`\nO ticket foi criado nesse canal de texto, para excluir use \
                        `/ticket delete-id id: {id}` ou clique no botão abaixo que terá o mesmo efeito.",
                    ))
                    .to_owned(),
            )
            .set_components(
                CreateActionRow::default()
                    .add_button(
                        CreateButton::default()
                            .style(ButtonStyle::Primary)
                            .label("Cancelar")
                            .emoji(ReactionType::Unicode("❌".to_owned()))
                            .custom_id("ticket-delete/".to_owned() + &id)
                            .to_owned(),
                    )
                    .to_components(),
            )
            .to_owned();

        send_message(&http, msg, channel.id.0)
            .await
            .or_else(|e| Err(BotError::Serenity(e)))?;

        if let Err(e) = self.ticket_repo.create(data).await {
            http.as_ref()
                .delete_channel(channel.id.0)
                .await
                .or_else(|e| Err(BotError::Serenity(e)))?;

            return Err(e);
        }

        Ok(InteractionResponse::default()
            .set_kind(InteractionResponseType::ChannelMessageWithSource)
            .set_data(
                CreateInteractionResponseData::default()
                    .content(format!("Seu ticket foi criado, <@{user_id}>"))
                    .set_components(
                        CreateActionRow::default()
                            .add_button(
                                CreateButton::default()
                                    .style(ButtonStyle::Link)
                                    .label("Ir")
                                    .url(format!(
                                        "https://discord.com/channels/{guild_id}/{}",
                                        channel.id.0
                                    ))
                                    .emoji(ReactionType::Unicode("🚀".to_owned()))
                                    .to_owned(),
                            )
                            .to_components(),
                    )
                    .ephemeral(true),
            )
            .to_owned())
    }

    pub async fn handle_delete_ticket_by_id(
        &self,
        http: impl AsRef<Http>,
        options: &Vec<CommandDataOption>,
        user_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let id = get_option(options, "id")
            .ok_or(BotError::OptionNotProvided("id"))?
            .value
            .ok_or(BotError::InvalidOption("id"))?
            .as_str()
            .ok_or(BotError::InvalidOption("id"))?
            .to_owned();

        let ticket = self.ticket_repo.get(id.clone()).await?;

        if (ticket.user_id as u64) != user_id {
            return Err(BotError::TicketDeletionDenied(id));
        }

        match http.as_ref().delete_channel(ticket.channel_id as u64).await {
            Ok(c) => log::info!("Channel {} deleted", c.id().0),
            Err(e) => log::warn!("Failed to delete channel: {e}"),
        }

        if let Err(e) = self.ticket_repo.delete(id).await {
            match e {
                BotError::TicketNotFound => Err(e),
                _ => {
                    log::error!("Failed to delete ticket: {e}");
                    Err(e)
                }
            }
        } else {
            Ok(())
        }?;

        Ok(InteractionResponse::default()
            .set_kind(InteractionResponseType::ChannelMessageWithSource)
            .set_data(
                CreateInteractionResponseData::default()
                    .content(format!("Ticket deletado com sucesso <@{user_id}>")),
            )
            .to_owned())
    }
}
