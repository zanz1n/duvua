use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData, CustomCommandType},
};
use reqwest::Client;
use serenity::{
    builder::CreateApplicationCommand,
    model::prelude::application_command::ApplicationCommandInteraction, prelude::Context,
};
use std::{sync::Mutex, time::Instant};

pub struct PingCommand {
    data: &'static CommandHandlerData,
    ping_vec: Mutex<Vec<u64>>,
    ping_vec_len: Mutex<usize>,
    computed_average: Mutex<Option<u64>>,
    client: Client,
}

impl PingCommand {
    pub fn new() -> Self {
        Self {
            data: Box::leak(Box::new(build_data())),
            ping_vec: Mutex::new(Vec::new()),
            ping_vec_len: Mutex::new(0),
            client: Client::new(),
            computed_average: Mutex::new(None),
        }
    }

    fn get_pre_computed(&self) -> Option<u64> {
        let lock = self.computed_average.lock().unwrap();
        let ping = lock.clone();
        drop(lock);
        ping
    }

    fn get_len(&self) -> usize {
        let lock = self.ping_vec_len.lock().unwrap();
        let len = lock.clone();
        drop(lock);
        len
    }

    fn append_len(&self) {
        let mut lock = self.ping_vec_len.lock().unwrap();
        *lock = *lock + 1;
        drop(lock)
    }
}

#[async_trait]
impl CommandHandler for PingCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let average: u64;

        if let Some(ping) = self.get_pre_computed() {
            log::info!("Ping already computed: {ping:?}");

            average = ping
        } else {
            if self.get_len() < 10 {
                let start = Instant::now();
                self.client
                    .get("https://discord.com/api/v10/")
                    .send()
                    .await
                    .or_else(|e| {
                        log::error!("Failed send ping request: {e}");
                        Err(BotError::SomethingWentWrong)
                    })?;
                let took = (Instant::now() - start).as_millis().try_into().unwrap();

                let mut lock = self.ping_vec.lock().unwrap();
                lock.push(took);

                log::info!("Added ping mid calc: {lock:?}");

                self.append_len();

                drop(lock);

                average = took;
            } else {
                let vec_lock = self.ping_vec.lock().unwrap();
                let len: u64 = vec_lock.len().try_into().unwrap();
                let mut sum = 0;

                for n in vec_lock.iter() {
                    sum += *n;
                }
                drop(vec_lock);

                sum = sum / len;
                average = sum;

                *self.computed_average.lock().unwrap() = Some(sum);

                log::info!("Ping computed: {sum:?}");
            }
        }

        InteractionResponse::with_content(format!("🏓 **Pong!**\n📡 Ping do bot: {average}ms"))
            .respond(&ctx.http, interaction.id.0, &interaction.token)
            .await
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
        kind: CustomCommandType::Info,
    }
}

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("ping")
        .description("Responde com pong e mostra o ping do bot")
        .description_localized("en-US", "Replies with pong and shows the bot latency")
        .to_owned()
}
