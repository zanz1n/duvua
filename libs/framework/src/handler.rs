use crate::{builder::hashmap_to_json_map, errors::BotError};
use async_trait::async_trait;
use deadpool_redis::{
    redis::{AsyncCommands, RedisError},
    Pool,
};
use serde::{Deserialize, Serialize};
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

#[derive(Debug, Clone, Deserialize)]
pub enum CustomCommandType {
    Fun,
    Info,
    Moderation,
    Config,
    Music,
    Money,
    Level,
    Utility,
}

impl Serialize for CustomCommandType {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_u8(match self {
            Self::Fun => 0,
            Self::Info => 1,
            Self::Moderation => 2,
            Self::Config => 3,
            Self::Music => 4,
            Self::Money => 5,
            Self::Level => 6,
            Self::Utility => 7,
        })
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommandPayload {
    pub data: Value,
    pub kind: CustomCommandType,
}

#[derive(Debug, Clone)]
pub struct CommandHandlerData {
    pub needs_defer: bool,
    pub command_data: Option<CreateApplicationCommand>,
    pub custom_id: Option<String>,
    pub kind: CustomCommandType,
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

        let name = if let Some(command_data) = &info.command_data {
            command_data
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

    pub fn get_pubsub_payload(&self) -> Vec<CommandPayload> {
        let mut array: Vec<CommandPayload> = Vec::new();

        for (_, v) in self.mp.iter() {
            let data = v.get_data();

            if let Some(command_data) = &data.command_data {
                array.push(CommandPayload {
                    data: Value::Object(hashmap_to_json_map(command_data.0.clone())),
                    kind: data.kind.clone(),
                })
            }
        }

        array
    }

    pub fn get_application_commands_data(&self) -> Vec<CreateApplicationCommand> {
        let mut array: Vec<CreateApplicationCommand> = Vec::new();

        for (_, v) in self.mp.iter() {
            let data = v.get_data();

            if let Some(command_data) = &data.command_data {
                array.push(command_data.clone())
            }
        }

        array
    }

    pub async fn post_commands(&self, http: impl AsRef<Http>) {
        let mut success_len = 0;
        let init_len: usize;

        if let Some(client) = &self.redis_client {
            let commands_data = self.get_pubsub_payload();
            init_len = commands_data.len();

            let mut conn = match client.get().await {
                Ok(v) => v,
                Err(e) => {
                    log::error!(target: "handler", "Failed to get redis connection: {e}");
                    return;
                }
            };

            for command in commands_data {
                let encoded = match serde_json::to_string(&command) {
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
            let commands_data = self.get_application_commands_data();
            init_len = commands_data.len();

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
            if data.command_data.is_some() {
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

            if data.custom_id.is_some() {
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
