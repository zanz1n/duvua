use super::ticket_shared::TicketSharedHandler;
use async_trait::async_trait;
use duvua_framework::{errors::BotError, handler::CommandHandler};
use serenity::{model::prelude::message_component::MessageComponentInteraction, prelude::Context};
use std::sync::Arc;

pub struct MessageComponentHandler {
    shared_handler: Arc<TicketSharedHandler>,
}

impl MessageComponentHandler {
    pub fn new(shared_handler: Arc<TicketSharedHandler>) -> Self {
        Self { shared_handler }
    }
}

#[async_trait]
impl CommandHandler for MessageComponentHandler {
    async fn handle_component(
        &self,
        ctx: &Context,
        interaction: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        let custom_id = interaction.data.custom_id.clone();

        if custom_id.starts_with("ticket-delete/") && custom_id.len() > 14 {
            let (_, id) = custom_id.split_at(14);
            self.shared_handler
                .handle_delete_ticket_by_id(&ctx.http, id.to_owned(), interaction.user.id.0)
                .await?;
        }

        Ok(())
    }
}
