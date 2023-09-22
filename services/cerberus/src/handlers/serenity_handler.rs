use super::welcome_shared::WelcomeSharedHandler;
use async_trait::async_trait;
use duvua_cache::CacheRepository;
use duvua_framework::handler::Handler;
use serenity::{
    model::prelude::{Interaction, Member, Ready},
    prelude::{Context, EventHandler},
};
use std::sync::Arc;

pub struct SerenityHandler<C: CacheRepository> {
    framework_handler: Handler,
    shared_handler: Arc<WelcomeSharedHandler<C>>,
}

impl<C: CacheRepository> SerenityHandler<C> {
    pub fn new(framework_handler: Handler, shared_handler: Arc<WelcomeSharedHandler<C>>) -> Self {
        Self {
            framework_handler,
            shared_handler,
        }
    }
}

#[async_trait]
impl<C: CacheRepository> EventHandler for SerenityHandler<C> {
    async fn ready(&self, ctx: Context, info: Ready) {
        self.framework_handler.ready(ctx, info).await
    }

    async fn interaction_create(&self, ctx: Context, interaction: Interaction) {
        self.framework_handler
            .interaction_create(ctx, interaction)
            .await
    }

    async fn guild_member_addition(&self, ctx: Context, member: Member) {
        _ = self
            .shared_handler
            .trigger(&ctx.http, member.guild_id.0, &member)
            .await;
    }
}
