use crate::errors::BotError;
use async_trait::async_trait;
use serenity::{
    builder::CreateApplicationCommand,
    model::prelude::{
        application_command::ApplicationCommandInteraction,
        message_component::MessageComponentInteraction,
    },
    prelude::Context,
};

#[derive(Debug, Clone, Default)]
pub struct CommandHandlerData {
    pub accepts_message_component: bool,
    pub accepts_application_command: bool,
    pub needs_defer: bool,
    pub command_data: Option<CreateApplicationCommand>,
    pub custom_id: Option<String>,
}

impl AsRef<CommandHandlerData> for CommandHandlerData {
    fn as_ref(&self) -> &CommandHandlerData {
        self
    }
}

#[async_trait]
pub trait CommandHandler: Send + Sync {
    async fn handle_command(
        &self,
        _ctx: &Context,
        _interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
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
        todo!()
    }
}
