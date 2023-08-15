use crate::builder::interaction_response::InteractionResponse;
use serenity::{
    model::prelude::{
        application_command::ApplicationCommandInteraction,
        message_component::MessageComponentInteraction,
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
    #[error("Option '{0}' not provided")]
    OptionNotProvided(&'static str),
    #[error("Option '{0}' is invalid provided")]
    InvalidOption(&'static str),
    #[error("Something went wrong")]
    SomethingWentWrong,
    #[error("You can not delete a ticket that is not yours")]
    TicketDeletionDenied(String),
    #[error("Redis error")]
    RedisError,
    #[error("Failed to deserialize cache")]
    CacheDeserializeError,
    #[error("Failed to serialize cache")]
    CacheSerializeError,
    #[error("Permission denied to issue this command")]
    CommandPermissionDenied,
    #[error("Failed to send ticket permanent message")]
    FailedToSendChannelMessage,
    #[error("The provided channel is invalid")]
    InvalidChannelProvided,
    #[error("Failed to fetch user avatar")]
    UserAvatarFetchFailed,
}

impl BotError {
    pub fn get_message(&self) -> String {
        if let Self::OptionNotProvided(s) = self {
            return format!("Opção '{s}' não foi fornecida");
        } else if let Self::InvalidOption(s) = self {
            return format!("Opção '{s}' é inválida");
        } else if let Self::TicketDeletionDenied(s) = self {
            return format!(
                "Você não pode deletar um ticket que não é seu! Caso seja \
                administrador, use o comando `/ticketadmin delete id: {s}`"
            );
        }

        match self {
            Self::UserNotFound => "Não foi possível encontrar o usuário",
            Self::UserAlreadyExists => "O usuário já existe",
            Self::TicketNotFound => "Não foi possível achar nenhum ticket",
            Self::CommandIssuedOutOfGuild => "Esse comando só pode ser usado dentro de um servidor",
            Self::GuildNotPermitTickets => "Tickets não estão habilidatos nesse servidor",
            Self::OnlyOneTicketAllowed => "O servidor só permite a criação de um ticket por membro",
            Self::CommandPermissionDenied => "Você não tem permissão para usar esse comando!",
            Self::FailedToSendChannelMessage => "Não foi possível enviar a mensagem no canal de texto",
            Self::InvalidChannelProvided => "O canal fornecido é inválido",
            Self::UserAvatarFetchFailed => "Não foi possível procurar o avatar do usuário",
            e => {
                log::error!(target: "framework_errors", "Unhandled command error: {}", e.to_string());
                "🤖 Algo deu errado!"
            }
        }
        .to_owned()
    }

    #[inline]
    pub fn get_response(&self, _defered: bool) -> InteractionResponse<'_> {
        InteractionResponse::with_content(self.get_message())
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
