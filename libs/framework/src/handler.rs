use crate::errors::BotError;
use async_trait::async_trait;
use deadpool_redis::{
    redis::{AsyncCommands, RedisError},
    Pool,
};
use serde_json::Value;
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommands},
    http::Http,
    model::prelude::{
        application_command::ApplicationCommandInteraction,
        message_component::MessageComponentInteraction, Interaction, Ready,
    },
    prelude::{Context, EventHandler},
};
use std::{collections::HashMap, sync::Arc, time::Instant};

#[derive(Debug, Clone, Default)]
pub struct CommandHandlerData {
    pub accepts_message_component: bool,
    pub accepts_application_command: bool,
    pub needs_defer: bool,
    pub command_data: Option<CreateApplicationCommand>,
    pub custom_id: Option<String>,
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

pub enum ComponentHandlerMode {
    FallBack,
    Overwrite,
}

pub struct Handler {
    mp: HashMap<String, Box<dyn CommandHandler>>,
    component_handler: Option<Box<dyn CommandHandler>>,
    component_handler_mode: Option<ComponentHandlerMode>,
    creation_instant: Instant,
    redis_client: Option<Arc<Pool>>,
}

impl Handler {
    pub fn new(cmds_pubsub: Option<Arc<Pool>>) -> Self {
        Self {
            creation_instant: Instant::now(),
            mp: HashMap::new(),
            component_handler: None,
            component_handler_mode: None,
            redis_client: cmds_pubsub,
        }
    }

    pub fn set_component_handler<H: CommandHandler + 'static>(
        &mut self,
        handler: H,
        fallback: bool,
    ) -> &mut Self {
        self.component_handler = Some(Box::new(handler));
        self.component_handler_mode = Some(if fallback {
            ComponentHandlerMode::FallBack
        } else {
            ComponentHandlerMode::Overwrite
        });
        self
    }

    pub fn add_handler<H: CommandHandler + 'static>(&mut self, handler: H) -> &mut Self {
        let info = handler.get_data();

        let name = if info.accepts_application_command {
            info.command_data
                .as_ref()
                .unwrap()
                .0
                .get("name")
                .unwrap()
                .as_str()
                .unwrap()
                .to_owned()
        } else {
            info.custom_id.clone().unwrap()
        };

        self.mp.insert(name, Box::new(handler));

        self
    }

    pub fn get_application_commands_data(&self) -> Vec<CreateApplicationCommand> {
        let mut array: Vec<CreateApplicationCommand> = Vec::new();

        for (_, v) in self.mp.iter() {
            let data = v.get_data();

            if data.accepts_application_command {
                array.push(data.command_data.clone().unwrap())
            }
        }

        array
    }

    pub async fn post_commands(&self, http: impl AsRef<Http>) {
        let commands_data = self.get_application_commands_data();
        let init_len = commands_data.len();

        let mut success_len = 0;

        if let Some(client) = &self.redis_client {
            let mut conn = match client.get().await {
                Ok(v) => v,
                Err(e) => {
                    log::error!(target: "handler", "Failed to get redis connection: {e}");
                    return;
                }
            };

            for command in commands_data {
                let encoded = match serde_json::to_string(&command.0) {
                    Ok(v) => v,
                    Err(e) => {
                        log::error!(target: "handler", "Failed to serialize application command: {e}");
                        continue;
                    }
                };

                let result: Result<i32, RedisError> = conn.publish("commands", encoded).await;
                match result {
                    Ok(_) => {
                        success_len += 1;
                    }
                    Err(e) => {
                        log::error!(target: "handler", "Failed to publish command message: {e}");
                    }
                }
            }
        } else {
            let mut create_application_commands = CreateApplicationCommands::default();
            create_application_commands.set_application_commands(commands_data);

            let result = http
                .as_ref()
                .create_global_application_commands(&Value::from(create_application_commands.0))
                .await;

            match result {
                Ok(v) => {
                    success_len = v.len();
                }
                Err(e) => {
                    log::error!(target: "handler", "Failed to post application commands: {e}");
                }
            }
        }

        log::info!(target: "handler", "Posted {success_len}/{init_len} commands");
    }

    pub async fn on_application_command(&self, ctx: Context, i: ApplicationCommandInteraction) {
        if let Some(cmd) = self.mp.get(&i.data.name) {
            let data = cmd.get_data();

            if data.needs_defer {
                if let Err(e) = i.defer(&ctx.http).await {
                    log::error!(target: "handler", "Failed to defer application command: {e}");
                    return;
                }
            }
            if data.accepts_application_command {
                let start = Instant::now();

                if let Err(e) = cmd.handle_command(&ctx, &i).await {
                    e.respond_application_command(&ctx, &i, data.needs_defer)
                        .await;
                }

                log::info!(target: "handler",
                    "Command handler executed in {}ms",
                    (Instant::now() - start).as_millis()
                )
            }
        }
    }

    pub async fn on_message_component(&self, ctx: Context, i: MessageComponentInteraction) {
        let fallback: bool;

        if let Some(mode) = self.component_handler_mode.as_ref() {
            if let ComponentHandlerMode::Overwrite = mode {
                _ = self
                    .component_handler
                    .as_ref()
                    .unwrap()
                    .handle_component(&ctx, &i)
                    .await;
                return;
            } else {
                fallback = true
            }
        } else {
            fallback = false
        }

        if let Some(cmd) = self.mp.get(&i.data.custom_id) {
            let data = cmd.get_data();

            if data.needs_defer {
                if let Err(e) = i.defer(&ctx.http).await {
                    log::error!(target: "handler", "Failed to defer message component: {e}");
                    return;
                }
            }

            if data.accepts_message_component {
                let start = Instant::now();

                if let Err(e) = cmd.handle_component(&ctx, &i).await {
                    e.respond_message_component(&ctx, &i, data.needs_defer)
                        .await;
                }

                log::info!(target: "handler",
                    "Component handler executed, took {}ms",
                    (Instant::now() - start).as_millis()
                )
            }
        } else if fallback {
            _ = self
                .component_handler
                .as_ref()
                .unwrap()
                .handle_component(&ctx, &i)
                .await;
        }
    }
}

#[async_trait]
impl EventHandler for Handler {
    async fn ready(&self, ctx: Context, info: Ready) {
        log::info!(target: "handler",
            "Logged in as {}. Initialization took {}ms",
            info.user.name,
            (Instant::now() - self.creation_instant).as_millis(),
        );

        self.post_commands(ctx.http.as_ref()).await;
    }

    async fn interaction_create(&self, ctx: Context, interaction: Interaction) {
        match interaction {
            Interaction::ApplicationCommand(i) => {
                self.on_application_command(ctx, i).await;
            }
            Interaction::MessageComponent(i) => {
                self.on_message_component(ctx, i).await;
            }
            _ => {}
        }
    }
}