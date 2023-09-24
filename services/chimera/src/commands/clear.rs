use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
    utils::get_option,
};
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};
use std::str::FromStr;

macro_rules! max_delete_size {
    () => {
        100
    };
}

pub struct ClearCommand {
    data: &'static CommandHandlerData,
}

impl ClearCommand {
    pub fn new() -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
        }
    }
}

#[async_trait]
impl CommandHandler for ClearCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let delete_limit = get_option(&interaction.data.options, "amount")
            .ok_or(BotError::OptionNotProvided("amount"))?
            .value
            .ok_or(BotError::InvalidOption("amount"))?
            .as_u64()
            .ok_or(BotError::InvalidOption("amount"))?;

        if delete_limit > max_delete_size!() || delete_limit < 1 {
            return Err(BotError::IntegerOptionOutOfRange {
                name: "amount",
                min: 1,
                max: max_delete_size!(),
            });
        }

        let delete_user = if let Some(v) = get_option(&interaction.data.options, "user") {
            match v.value {
                Some(v) => match v.as_str() {
                    Some(v) => Some(u64::from_str(v).or(Err(BotError::InvalidOption("user")))?),
                    None => None,
                },
                None => None,
            }
        } else {
            None
        };

        let skip_bots = match get_option(&interaction.data.options, "skip_bots") {
            Some(v) => v
                .value
                .ok_or(BotError::InvalidOption("skip_bots"))?
                .as_bool()
                .ok_or(BotError::InvalidOption("skip_bots"))?,
            None => false,
        };

        let channel_id = interaction.channel_id.0;
        let msgs = interaction
            .channel_id
            .messages(&ctx, |query| query.limit((delete_limit * 2) + 6))
            .await
            .map_err(|e| {
                log::error!("Failed to fetch messages of channel {channel_id}: {e}");
                BotError::ChannelMessagesFetchFailed(channel_id)
            })?;

        if msgs.len() == 0 {
            return Err(BotError::FilteredEmptyMessageSet);
        }

        let mut msgs = msgs
            .iter()
            .filter(|msg| {
                if let Some(user_id) = delete_user {
                    if msg.author.id.0 != user_id {
                        return false;
                    }
                }

                if skip_bots && msg.author.bot {
                    return false;
                }

                true
            })
            .map(|msg| msg.id)
            .collect::<Vec<_>>();

        let len = msgs.len();
        if len == 0 {
            return Err(BotError::FilteredEmptyMessageSet);
        }
        msgs.truncate(delete_limit as usize);

        interaction
            .channel_id
            .delete_messages(&ctx, msgs)
            .await
            .map_err(|e| {
                log::error!("Failed to delete messages of channel {channel_id}: {e}");
                BotError::MessageBulkDeleteFailed(channel_id)
            })?;

        InteractionResponse::with_content(format!(
            "{len} mensagens excluídas do canal de texto <#{channel_id}>"
        ))
        .respond(&ctx, interaction.id.0, &interaction.token)
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
        command_data: Some(build_data_command()),
        custom_id: None,
        needs_defer: false,
        kind: CustomCommandType::Moderation,
    }
}

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("clear")
        .description("Limpa um número de mensagens no chat")
        .description_localized("en-US", "Clears a certain number of messages from chat")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::Integer)
                .name("amount")
                .description("A quantidade de mensagens que deseja limpar (max 100)")
                .description_localized("en-US", "The amount of messages to clear (max 100)")
                .required(true)
                // .min_int_value(1)
                // .max_int_value(max_delete_size!())
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::User)
                .name("user")
                .description("Filtra para excluir mensagens apenas de um usuário")
                .description_localized(
                    "en-US",
                    "Deletes messages only if they were sent by this user",
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::Boolean)
                .name("skip_bots")
                .description("Pula mensagens enviadas por bots")
                .description_localized("en-US", "Skips messages sent by bots")
                .to_owned(),
        )
        .to_owned()
}
