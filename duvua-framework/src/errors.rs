use crate::builder::interaction_response::InteractionResponse;
use serenity::{
    builder::{CreateComponents, CreateInteractionResponseData},
    model::prelude::{
        application_command::ApplicationCommandInteraction,
        message_component::MessageComponentInteraction, InteractionResponseType,
    },
    prelude::Context,
};

#[derive(thiserror::Error, Debug)]
pub enum BotError {
    #[error("Serenity error: {0}")]
    Serenity(serenity::Error),
    #[error("User could not be found")]
    UserNotFound,
    #[error("Something went wrong while performing a query")]
    Query,
    #[error("User already exists")]
    UserAlreadyExists,
    #[error("Mongo db error")]
    MongoDbError,
    #[error("Guild could not be found")]
    TicketGuildNotFound,
    #[error("Guild already exists")]
    TicketGuildAlreadyExists,
    #[error("Invalid mongodb objectId")]
    InvalidMongoDbObjectId,
    #[error("Ticket could not be found")]
    TicketNotFound,
    #[error("This command can only be issued inside a guild")]
    CommandIssuedOutOfGuild,
    #[error("Guild not permit tickets")]
    GuildNotPermitTickets,
    #[error("This guild only allows one ticket per member")]
    OnlyOneTicketAllowed,
    #[error("Option not '{0}' provided")]
    OptionNotProvided(&'static str),
    #[error("Something went wrong")]
    SomethingWentWrong,
}

impl BotError {
    pub fn get_message(&self) -> String {
        if let Self::OptionNotProvided(s) = self {
            return format!("Opção '{s}' não foi fornecida");
        }

        match self {
            Self::UserNotFound => "Não foi possível encontrar o usuário",
            Self::UserAlreadyExists => "O usuário já existe",
            Self::TicketNotFound => "Não foi possível achar nenhum ticket",
            Self::CommandIssuedOutOfGuild => "Esse comando só pode ser usado dentro de um servidor",
            Self::GuildNotPermitTickets => "Ticket não estão habilidatos nesse servidor",
            Self::OnlyOneTicketAllowed => "O servidor só permite a criação de um ticket por membro",
            e => {
                log::error!(target: "framework_errors", "Unhandled command error: {}", e.to_string());
                "🤖 Algo deu errado!"
            }
        }
        .to_owned()
    }

    #[inline]
    pub fn get_response(&self, defered: bool) -> InteractionResponse<'_> {
        let mut response = InteractionResponse::default();
        response
            .set_kind(if defered {
                InteractionResponseType::UpdateMessage
            } else {
                InteractionResponseType::ChannelMessageWithSource
            })
            .set_data(
                CreateInteractionResponseData::default()
                    .set_components(CreateComponents::default())
                    .set_embeds(Vec::new())
                    .content(self.get_message()),
            );

        response
    }

    pub async fn respond_message_component(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
        defered: bool,
    ) {
        if let &BotError::Serenity(e) = &self {
            log::error!("Serenity error: {e}");
        } else {
            _ = self
                .get_response(defered)
                .respond_message_component(ctx.http.as_ref(), interaction)
                .await;
        }
    }

    pub async fn respond_application_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
        defered: bool,
    ) {
        if let &BotError::Serenity(e) = &self {
            log::error!("Serenity error: {e}");
        } else {
            _ = self
                .get_response(defered)
                .respond_application_command(ctx.http.as_ref(), interaction)
                .await;
        }
    }
}
