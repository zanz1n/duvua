use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
};
use serenity::{
    builder::{CreateApplicationCommand, CreateInteractionResponseData},
    model::prelude::application_command::ApplicationCommandInteraction,
    prelude::Context,
};

pub struct PingCommand {
    data: &'static CommandHandlerData,
}

impl PingCommand {
    pub fn new() -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
        }
    }
}

#[async_trait]
impl CommandHandler for PingCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        InteractionResponse::default()
            .set_data(
                CreateInteractionResponseData::default()
                    .content(format!("🏓 **Pong!**\n📡 Ping do bot: {}ms", "120")),
            )
            .respond_application_command(ctx.http.as_ref(), interaction)
            .await
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        accepts_application_command: true,
        accepts_message_component: false,
        command_data: Some(build_data_command()),
        custom_id: None,
        needs_defer: false,
    }
}

#[inline]
fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("ping")
        .description("Replies with pong and shows the bot latency")
        .to_owned()
}
