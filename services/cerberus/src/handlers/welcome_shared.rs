use duvua_cache::{utils::get_or_store_guild, CacheRepository};
use duvua_framework::{builder::interaction_response::InteractionResponse, errors::BotError};
use duvua_repository::welcome::{Welcome, WelcomeRepository, WelcomeService, WelcomeType};
use serenity::{
    builder::{CreateEmbed, CreateMessage},
    http::Http,
    json::{hashmap_to_json_map, Value},
    model::prelude::{Channel, ChannelType, Member},
};
use std::sync::Arc;

pub struct WelcomeSharedHandler<C: CacheRepository> {
    repository: Arc<WelcomeService>,
    cache: Arc<C>,
}

impl<C: CacheRepository> WelcomeSharedHandler<C> {
    pub fn new(repository: Arc<WelcomeService>, cache: Arc<C>) -> Self {
        Self { repository, cache }
    }

    pub async fn handle_set_message(
        &self,
        guild_id: u64,
        kind: WelcomeType,
        message: String,
    ) -> Result<InteractionResponse, BotError> {
        let guild_id = guild_id as i64;

        let msg = format!("Mensagem alterada para \"{message}\" com tipo '{kind}'");

        if self.repository.exists(guild_id).await? {
            self.repository
                .update_message(guild_id, message, kind)
                .await?;

            Ok(InteractionResponse::with_content(msg))
        } else {
            self.repository
                .create(Welcome {
                    id: guild_id,
                    message,
                    kind,
                    ..Default::default()
                })
                .await?;

            Ok(InteractionResponse::with_content(
                msg + ". Lembre se de habilitar usando `/welcomeadmin enable` e \
                configurar um canal de texto usando `/welcomeadmin set-channel`",
            ))
        }
    }

    pub async fn handle_set_channel(
        &self,
        http: impl AsRef<Http>,
        guild_id: u64,
        channel_id: u64,
    ) -> Result<InteractionResponse, BotError> {
        let channel = http.as_ref().get_channel(channel_id).await.or_else(|e| {
            log::error!("Failed to fetch guild channel for validation: {e}");
            Err(BotError::InvalidChannelProvided)
        })?;

        if let Channel::Guild(channel) = channel {
            if channel.kind != ChannelType::Text {
                return Err(BotError::InvalidChannelProvided);
            }
        } else {
            return Err(BotError::InvalidChannelProvided);
        }

        let guild_id = guild_id as i64;
        let channel_id = channel_id as i64;

        let msg =
            format!("Canal de texto das mensagens de boas vindas atualizado para <#{channel_id}>");

        if self.repository.exists(guild_id).await? {
            self.repository
                .update_channel_id(guild_id, Some(channel_id))
                .await?;

            Ok(InteractionResponse::with_content(msg))
        } else {
            self.repository
                .create(Welcome {
                    id: guild_id,
                    channel_id: Some(channel_id),
                    ..Default::default()
                })
                .await?;

            Ok(InteractionResponse::with_content(
                msg + ". Lembre-se de habilitar as mensagens usando `/welcomeadmin enable`",
            ))
        }
    }

    pub async fn handle_set_enabled(
        &self,
        guild_id: u64,
        enabled: bool,
    ) -> Result<InteractionResponse, BotError> {
        let guild_id = guild_id as i64;

        let msg = "Mensagem de boas vindas ".to_owned()
            + if enabled {
                "habilitada"
            } else {
                "desabilitada"
            };

        if self.repository.exists(guild_id).await? {
            self.repository.update_enabled(guild_id, enabled).await?;

            Ok(InteractionResponse::with_content(msg))
        } else {
            self.repository
                .create(Welcome {
                    id: guild_id,
                    enabled,
                    ..Default::default()
                })
                .await?;

            Ok(InteractionResponse::with_content(
                msg + ". Lembre-se de usar `/welcomeadmin set-channel`",
            ))
        }
    }

    pub async fn trigger(
        &self,
        http: &Arc<Http>,
        guild_id: u64,
        member: &Member,
    ) -> Result<bool, BotError> {
        let welcome = match self.repository.get_by_id(guild_id as i64).await? {
            Some(v) => v,
            None => return Ok(false),
        };

        let channel_id = match welcome.channel_id {
            Some(v) => v,
            None => return Ok(false),
        };

        let mut text_message = welcome
            .message
            .replace("{{USER}}", &format!("<@{}>", member.user.id.0));

        if text_message.contains("{{GUILD}}") {
            let guild = get_or_store_guild(http, &self.cache, guild_id).await?;
            text_message = text_message.replace("{{GUILD}}", &guild.name);
        }

        let mut message = CreateMessage::default();

        match welcome.kind {
            WelcomeType::Embed => {
                let mut embed = CreateEmbed::default();
                embed.title("Bem vindo(a)!");
                embed.description(text_message);

                message.set_embed(embed);
            }
            WelcomeType::Message => {
                message.content(text_message);
            }
            _ => {}
        }

        let http = http.clone();

        tokio::spawn(async move {
            let result = http
                .send_message(
                    channel_id as u64,
                    &Value::Object(hashmap_to_json_map(message.0)),
                )
                .await;

            match result {
                Ok(_) => {
                    log::info!("Welcome message sent on channel {channel_id}");
                }
                Err(e) => {
                    log::error!("Failed to send message on channel {channel_id}: {e}");
                }
            }
        });

        Ok(true)
    }
}
