use crate::repository::{
    guild::{Guild, GuildRepository},
    ticket::{CreateTicketData, TicketRepository},
};
use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_sub_command,
};
use serenity::{
    builder::{
        CreateActionRow, CreateApplicationCommand, CreateApplicationCommandOption, CreateButton,
        CreateChannel, CreateComponents, CreateInteractionResponseData,
    },
    http::Http,
    json::hashmap_to_json_map,
    model::{
        prelude::{
            application_command::ApplicationCommandInteraction, command::CommandOptionType,
            component::ButtonStyle, message_component::MessageComponentInteraction, ChannelType,
            InteractionResponseType, PermissionOverwrite, PermissionOverwriteType, ReactionType,
            RoleId, UserId,
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

        let channel = CreateChannel::default()
            .kind(ChannelType::Text)
            .name(username + "-" + data.id.to_hex().as_str())
            .permissions(permissions)
            .to_owned();

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
                        CreateComponents::default()
                            .add_action_row(
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
                                    .to_owned(),
                            )
                            .to_owned(),
                    )
                    .ephemeral(true),
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

        let sub_command = get_sub_command(&interaction.data.options)
            .ok_or(BotError::SomethingWentWrong)?
            .name;
        let sub_command = sub_command.as_str();

        if sub_command == "create" {
            let res = self
                .handle_create_ticket(
                    &ctx.http,
                    guild,
                    guild_id,
                    user_id,
                    interaction.user.name.clone(),
                )
                .await?;

            res.respond(&ctx.http, interaction.id.0, &interaction.token)
                .await?;
        } else if sub_command == "delete" {
        }

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
                .name("delete")
                .description("Deleta seus tickets caso você tenha algum")
                .description_localized("en-US", "Deletes a ticket in case you have one open")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::String)
                        .name("id")
                        .description("O id do ticket que deseja deletar")
                        .description_localized("en-US", "The id of the ticket you want to delete")
                        .to_owned(),
                )
                .to_owned(),
        )
        .to_owned()
}
