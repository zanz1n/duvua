use async_trait::async_trait;
use duvua_framework::{
    builder::interaction_response::InteractionResponse,
    errors::BotError,
    handler::{CommandHandler, CommandHandlerData},
    utils::{get_option, get_sub_command},
};
use reqwest::Client;
use serenity::{
    builder::{CreateApplicationCommand, CreateApplicationCommandOption},
    model::prelude::{
        application_command::ApplicationCommandInteraction, command::CommandOptionType,
    },
    prelude::Context,
};

pub struct FactsCommand {
    client: Client,
    data: &'static CommandHandlerData,
}

impl FactsCommand {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
            data: Box::leak(Box::new(build_data())),
        }
    }
}

#[async_trait]
impl CommandHandler for FactsCommand {
    async fn handle_command(
        &self,
        ctx: &Context,
        interaction: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        let sub_command = get_sub_command(&interaction.data.options)
            .ok_or(BotError::InvalidOption("sub-command"))?;

        let number = get_option(&sub_command.options, "number")
            .ok_or(BotError::OptionNotProvided("number"))?
            .value
            .ok_or(BotError::InvalidOption("number"))?
            .as_i64()
            .ok_or(BotError::InvalidOption("number"))?;

        let res = match sub_command.name.as_str() {
            "year" => text_req(&self.client, format!("http://numbersapi.com/{number}")).await?,
            "number" => text_req(&self.client, format!("http://numbersapi.com/{number}")).await?,
            _ => return Err(BotError::InvalidOption("sub-command")),
        };

        InteractionResponse::with_content(res.to_string())
            .respond(&ctx.http, interaction.id.0, &interaction.token)
            .await
    }

    fn get_data(&self) -> &'static CommandHandlerData {
        self.data
    }
}

async fn text_req(http: &Client, url: String) -> Result<String, BotError> {
    http.get(url)
        .send()
        .await
        .or_else(|e| {
            log::error!("Failed to make http request: {e}");
            Err(BotError::SomethingWentWrong)
        })?
        .text()
        .await
        .or_else(|e| {
            log::error!("Failed decode request body: {e}");
            Err(BotError::SomethingWentWrong)
        })
}

#[inline]
fn build_data() -> CommandHandlerData {
    CommandHandlerData {
        accepts_message_component: false,
        accepts_application_command: true,
        needs_defer: false,
        command_data: Some(build_data_command()),
        custom_id: None,
    }
}

fn build_data_command() -> CreateApplicationCommand {
    CreateApplicationCommand::default()
        .name("facts")
        .description("Exibe curiosidades sobre números")
        .description_localized("en-US", "Shows facts about numbers")
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("year")
                .description("Exibe curiosidades sobre um ano")
                .description_localized("en-US", "Show facts about some year")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::Integer)
                        .name("number")
                        .description("O número que deseja pesquisar")
                        .description_localized("en-US", "The number you want to search")
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .add_option(
            CreateApplicationCommandOption::default()
                .kind(CommandOptionType::SubCommand)
                .name("number")
                .description("Exibe curiosidades sobre um número qualquer")
                .description_localized("en-US", "Show facts about some arbitrary number")
                .add_sub_option(
                    CreateApplicationCommandOption::default()
                        .kind(CommandOptionType::Integer)
                        .name("number")
                        .description("O número que deseja pesquisar")
                        .description_localized("en-US", "The number you want to search")
                        .required(true)
                        .to_owned(),
                )
                .to_owned(),
        )
        .to_owned()
}
