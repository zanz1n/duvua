use super::ticket_shared::TicketSharedHandler;
use crate::repository::guild::GuildRepository;
use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::get_sub_command,
};
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
        message_component::MessageComponentInteraction,
    },
    prelude::Context,
};
use std::sync::Arc;

pub struct TicketCommandHandler {
    guild_repo: Arc<dyn GuildRepository>,
    shared_handler: Arc<TicketSharedHandler>,
    data: &'static CommandHandlerData,
}

impl TicketCommandHandler {
    pub fn new(
        guild_repo: Arc<dyn GuildRepository>,
        shared_handler: Arc<TicketSharedHandler>,
    ) -> Self {
        Self {
            shared_handler,
            guild_repo,
            data: Box::leak(Box::new(build_data())),
        }
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

        let sub_command =
            get_sub_command(&interaction.data.options).ok_or(BotError::SomethingWentWrong)?;
        let sub_command_str = sub_command.name.as_str();

        let res: InteractionResponse<'_>;

        if sub_command_str == "create" {
            res = self
                .shared_handler
                .handle_create_ticket(
                    &ctx.http,
                    guild,
                    guild_id,
                    user_id,
                    interaction.user.name.clone(),
                )
                .await?;
        } else if sub_command_str == "delete-id" {
            res = self
                .shared_handler
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
