use crate::repository::{
    guild::GuildRepository, ticket::TicketRepository, ticket_shared::TicketSharedHandler,
};
use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::{get_option, get_sub_command, get_sub_command_group},
};
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    http::Http,
    model::prelude::{
        application_command::{ApplicationCommandInteraction, CommandDataOption},
        command::CommandOptionType,
        message_component::MessageComponentInteraction,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct TicketCommandHandler {
    guild_repo: Arc<dyn GuildRepository>,
    ticket_repo: Arc<dyn TicketRepository>,
    shared_handler: Arc<TicketSharedHandler>,
    data: &'static CommandHandlerData,
}

impl TicketCommandHandler {
    pub fn new(
        guild_repo: Arc<dyn GuildRepository>,
        ticket_repo: Arc<dyn TicketRepository>,
        shared_handler: Arc<TicketSharedHandler>,
    ) -> Self {
        Self {
            shared_handler,
            guild_repo,
            ticket_repo,
            data: Box::leak(Box::new(build_data())),
        }
    }

    pub async fn handle_delete_ticket_by_options(
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

        self.shared_handler
            .handle_delete_ticket_by_id(http, id, user_id)
            .await
    }

    pub async fn handle_delete_all(
        &self,
        http: &Arc<Http>,
        guild_id: u64,
        user_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let tickets = self.ticket_repo.get_by_member(guild_id, user_id, 6).await?;

        if tickets.len() > 5 {
            return Ok(InteractionResponse::with_content(
                "Você tem mais de 5 tickets criados, por favor exclua \
                eles individualmente usando `/ticket delete by-id`",
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
impl CommandHandler for TicketCommandHandler {
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

        let res: InteractionResponse<'_>;

        if let Some(sub_command) = get_sub_command(&interaction.data.options) {
            match sub_command.name.as_str() {
                "create" => {
                    res = self
                        .shared_handler
                        .handle_create_ticket(
                            &ctx.http,
                            guild,
                            guild_id,
                            user_id,
                            interaction.user.name.clone(),
                        )
                        .await?
                }
                _ => return Err(BotError::InvalidOption("sub-command")),
            }
            return res
                .respond(&ctx.http, interaction.id.0, &interaction.token)
                .await;
        }

        let sub_command_group =
            get_sub_command_group(&interaction.data.options).ok_or(BotError::SomethingWentWrong)?;

        if sub_command_group.name.as_str() == "delete" {
            let sub_command =
                get_sub_command(&sub_command_group.options).ok_or(BotError::SomethingWentWrong)?;

            match sub_command.name.as_str() {
                "by-id" => {
                    res = self
                        .handle_delete_ticket_by_options(&ctx.http, &sub_command.options, user_id)
                        .await?
                }
                "all" => {
                    res = self.handle_delete_all(&ctx.http, guild_id, user_id).await?;
                }
                _ => return Err(BotError::InvalidOption("sub-command")),
            }

            return res
                .respond(&ctx.http, interaction.id.0, &interaction.token)
                .await;
        }

        Ok(())
    }

    async fn handle_component(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        let guild_id = interaction
            .guild_id
            .ok_or(BotError::CommandIssuedOutOfGuild)?
            .0;
        let user_id = interaction.user.id.0;

        let guild = self.guild_repo.get_or_create(guild_id).await?;

        if !guild.enable_tickets {
            return Err(BotError::GuildNotPermitTickets);
        }

        self.shared_handler
            .handle_create_ticket(
                &ctx.http,
                guild,
                guild_id,
                user_id,
                interaction.user.name.clone(),
            )
            .await?
            .respond_message_component(&ctx.http, interaction)
            .await?;

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
        custom_id: Some("ticket".to_owned()),
        needs_defer: false,
    }
}

#[inline]
fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("ticket")
        .description("Comandos relacionados a tickets")
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
                .kind(CommandOptionType::SubCommandGroup)
                .name("delete")
                .description("Comandos para deletar tickets")
                .description_localized("en-US", "Commands to delete tickets")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::SubCommand)
                        .name("by-id")
                        .description("Deleta um ticket seu por id")
                        .description_localized("en-US", "Deletes one of your tickets by id")
                        .add_sub_option(
                            CreateApplicationCommandOption::default()
                                .kind(CommandOptionType::String)
                                .name("id")
                                .description("O id do ticket que deseja deletar")
                                .description_localized(
                                    "en-US",
                                    "The id of the ticket you want to delete",
                                )
                                .required(true)
                                .to_owned(),
                        )
                        .to_owned(),
                )
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::SubCommand)
                        .name("all")
                        .description("Deleta todos os seus tickets caso você tenha algum")
                        .description_localized("en-US", "Delete all your tickets if any")
                        .to_owned(),
                )
                .to_owned(),
        )
        .to_owned()
}
