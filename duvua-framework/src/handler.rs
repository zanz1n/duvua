use crate::models::CommandHandler;
use async_trait::async_trait;
use serenity::{
    http::Http,
    model::prelude::{command::Command, Interaction, Ready},
    prelude::{Context, EventHandler},
};
use std::{collections::HashMap, time::Instant};

pub struct Handler {
    mp: HashMap<String, Box<dyn CommandHandler>>,
    post_cmds_on_ready: bool,
}

impl Handler {
    pub fn new(post_cmds_on_ready: bool) -> Self {
        Self {
            mp: HashMap::new(),
            post_cmds_on_ready,
        }
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

    pub async fn post_commands(&self, http: impl AsRef<Http>) {
        let mut i = 0;

        let result = Command::set_global_application_commands(http, |cmds| {
            for (_, v) in self.mp.iter() {
                i += 1;
                let info = v.get_data();

                if info.accepts_application_command {
                    cmds.add_application_command(info.command_data.clone().unwrap());
                }
            }

            cmds
        })
        .await;

        match result {
            Ok(v) => {
                log::info!(target: "handler", "Posted {}/{i} commands", v.len());
            }
            Err(e) => {
                log::error!(target: "handler", "Failed to post application commands: {e}");
            }
        }
    }
}

#[async_trait]
impl EventHandler for Handler {
    async fn ready(&self, ctx: Context, info: Ready) {
        log::info!(target: "handler", "Logged in as {}", info.user.name);

        if self.post_cmds_on_ready {
            self.post_commands(ctx.http.as_ref()).await;
        }
    }

    async fn interaction_create(&self, ctx: Context, interaction: Interaction) {
        match interaction {
            Interaction::ApplicationCommand(i) => {
                if let Some(cmd) = self.mp.get(&i.data.name) {
                    let data = cmd.get_data();

                    if data.needs_defer {
                        if let Err(e) = i.defer(&ctx.http).await {
                            log::error!(target: "handler", "Failed to defer interaction: {e}");
                            return;
                        }
                    }
                    if data.accepts_application_command {
                        let start = Instant::now();

                        match cmd.handle_command(&ctx, &i).await {
                            Ok(_) => {}
                            Err(e) => {
                                e.respond_to(&ctx, &i, data.needs_defer).await;
                            }
                        };

                        log::info!(target: "handler", "Command handler executed in {}ms", (Instant::now() - start).as_millis())
                    }
                }
            }
            Interaction::MessageComponent(i) => {
                if let Some(cmd) = self.mp.get(&i.data.custom_id) {
                    let data = cmd.get_data();

                    if data.accepts_message_component {
                        let start = Instant::now();

                        match cmd.handle_component(&ctx, &i).await {
                            Ok(_) => {
                                log::info!(target: "handler", "Component handler executed, took {}ms", (Instant::now() - start).as_millis())
                            }
                            Err(e) => {
                                log::info!(target: "handler", "Command executed in {}ms with a error {e}", (Instant::now() - start).as_millis())
                            }
                        }
                    }
                }
            }
            _ => {}
        }
    }
}
