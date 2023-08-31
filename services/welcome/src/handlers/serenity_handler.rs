use async_trait::async_trait;
use duvua_framework::handler::Handler;
use serenity::{
    model::prelude::{Interaction, Member, Ready},
    prelude::{Context, EventHandler},
};

pub struct SerenityHandler {
    framework_handler: Handler,
}

impl SerenityHandler {
    pub fn new(framework_handler: Handler) -> Self {
        Self { framework_handler }
    }
}

#[async_trait]
impl EventHandler for SerenityHandler {
    async fn ready(&self, ctx: Context, info: Ready) {
        self.framework_handler.ready(ctx, info).await
    }

    async fn interaction_create(&self, ctx: Context, interaction: Interaction) {
        self.framework_handler
            .interaction_create(ctx, interaction)
            .await
    }

    async fn guild_member_addition(&self, _ctx: Context, member: Member) {
        log::info!("GuildMemberAdd {member}");
    }
}
