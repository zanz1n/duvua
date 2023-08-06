use crate::repository::{
    guild::{Guild, GuildRepository},
    ticket::{CreateTicketData, TicketRepository},
};
use async_trait::async_trait;
use duvua_framework::{
    builder::{button_action_row::CreateActionRow, interaction_response::InteractionResponse},
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::{get_option, get_sub_command, send_message},
};
use serenity::{
    builder::{
        CreateApplicationCommand, CreateApplicationCommandOption, CreateButton, CreateChannel,
        CreateEmbed, CreateInteractionResponseData, CreateMessage,
    },
    http::Http,
    json::hashmap_to_json_map,
    model::{
        prelude::{
            application_command::{ApplicationCommandInteraction, CommandDataOption},
            command::CommandOptionType,
            component::ButtonStyle,
            message_component::MessageComponentInteraction,
            ChannelType, InteractionResponseType, PermissionOverwrite, PermissionOverwriteType,
            ReactionType, RoleId, UserId,
        },
        Permissions,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct TicketCommand {
    ticket_repo: Arc<dyn TicketRepository>,
    guild_repo: Arc<dyn GuildRepository>,
    data: &'static CommandHandlerData,
}

impl TicketCommand {
    pub fn new(
        ticket_repo: Arc<dyn TicketRepository>,
        guild_repo: Arc<dyn GuildRepository>,
    ) -> Self {
        Self {
            ticket_repo,
            guild_repo,
            data: Box::leak(Box::new(build_data())),
        }
    }

    async fn handle_create_ticket(
        &self,
        http: impl AsRef<Http>,
        guild: Guild,
        guild_id: u64,
        user_id: u64,
        username: String,
        guild_channel_category: Option<u64>,
    ) -> Result<InteractionResponse, BotError> {
        if !guild.allow_multiple {
            let ticket = self.ticket_repo.get_by_member(guild_id, user_id).await;
            if ticket.is_ok() {
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

        if let Some(cat) = guild_channel_category {
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

    async fn handle_delete_ticket_by_id(
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

#[async_trait]
impl CommandHandler for TicketCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        if interaction.guild_id.is_none() {
            return Err(BotError::CommandIssuedOutOfGuild);
        }

        let guild_id = interaction.guild_id.unwrap().0;
        let user_id = interaction.user.id.0;

        let guild = self.guild_repo.get_or_create(guild_id).await?;

        if !guild.enable_tickets {
            return Err(BotError::GuildNotPermitTickets);
        }

        let sub_command =
            get_sub_command(&interaction.data.options).ok_or(BotError::SomethingWentWrong)?;
        let sub_command_str = sub_command.name.as_str();

        let res: InteractionResponse<'_>;

        if sub_command_str == "create" {
            let cat = guild.channel_category;

            res = self
                .handle_create_ticket(
                    &ctx.http,
                    guild,
                    guild_id,
                    user_id,
                    interaction.user.name.clone(),
                    cat,
                )
                .await?;
        } else if sub_command_str == "delete-id" {
            res = self
                .handle_delete_ticket_by_id(&ctx.http, &sub_command.options, user_id)
                .await?;
        } else {
            return Err(BotError::InvalidOption("sub-command"));
        }

        res.respond(&ctx.http, interaction.id.0, &interaction.token)
            .await?;

        Ok(())
    }

    async fn handle_component(
        &self,
        _ctx: &Context,
        _interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        Ok(())
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        accepts_application_command: true,
        accepts_message_component: true,
        command_data: Some(build_data_command()),
        custom_id: Some("ticket-create".to_owned()),
        needs_defer: false,
    }
}

#[inline]
fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("ticket")
        .description("Comandos para a criação de tickets")
        .description_localized("en-US", "Ticket realted commands")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("create")
                .description("Cria um ticket")
                .description_localized("en-US", "Creates a ticket")
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("delete-id")
                .description("Deleta um ticket seu por id")
                .description_localized("en-US", "Delete one your tickets by id")
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
                .name("delete")
                .description("Deleta seus tickets caso você tenha algum")
                .description_localized("en-US", "Delete a ticket in case you have one open")
                .to_owned(),
        )
        .to_owned()
}
