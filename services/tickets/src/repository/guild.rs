use async_trait::async_trait;
use duvua_framework::errors::BotError;
use mongodb::{
    bson::{doc, Document},
    error::ErrorKind,
    Collection,
};
use serde::{Deserialize, Serialize};
use std::ops::Deref;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Guild {
    #[serde(rename = "_id")]
    pub id: i64,
    pub enable_tickets: bool,
    pub allow_multiple: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateGuildData {
    pub enable_tickets: Option<bool>,
    pub allow_multiple: Option<bool>,
}

impl Guild {
    #[inline]
    pub fn new(id: u64) -> Self {
        Self {
            id: id as i64,
            enable_tickets: false,
            allow_multiple: false,
        }
    }
}

#[async_trait]
pub trait GuildRepository: Sync + Send {
    async fn get(&self, id: u64) -> Result<Guild, BotError>;
    async fn update(&self, id: u64, data: UpdateGuildData) -> Result<(), BotError>;
    async fn create(&self, id: u64) -> Result<Guild, BotError>;
    async fn get_or_create(&self, id: u64) -> Result<Guild, BotError>;
    async fn update_or_create(&self, id: u64, data: UpdateGuildData) -> Result<Guild, BotError>;
}

pub struct GuildService {
    col: Collection<Guild>,
}

impl GuildService {
    pub fn new(col: Collection<Guild>) -> Self {
        Self { col }
    }
}

#[async_trait]
impl GuildRepository for GuildService {
    async fn get(&self, id: u64) -> Result<Guild, BotError> {
        self.col
            .find_one(doc! {"_id": id as i64}, None)
            .await
            .or_else(|e| {
                log::error!(target: "mongo_errors", "{e}");
                Err(BotError::MongoDbError)
            })?
            .ok_or(BotError::TicketGuildNotFound)
    }

    async fn update(&self, id: u64, data: UpdateGuildData) -> Result<(), BotError> {
        let mut document = Document::new();

        if let Some(enable_tickets) = data.enable_tickets {
            document.insert("enable_tickets", enable_tickets);
        }
        if let Some(allow_multiple) = data.allow_multiple {
            document.insert("allow_multiple", allow_multiple);
        }

        let update_result = self
            .col
            .update_one(doc! {"_id": id as i64}, doc! {"$set": document}, None)
            .await
            .or_else(|e| {
                log::error!(target: "mongo_errors", "{e}");
                Err(BotError::MongoDbError)
            })?;

        if update_result.matched_count < 1 {
            return Err(BotError::TicketGuildNotFound);
        }

        Ok(())
    }

    async fn create(&self, id: u64) -> Result<Guild, BotError> {
        let guild = Guild::new(id);

        self.col.insert_one(&guild, None).await.or_else(|e| {
            Err(match e.kind.deref() {
                ErrorKind::Write(_) => BotError::TicketGuildAlreadyExists,
                _ => {
                    log::error!(target: "mongo_errors", "{e}");
                    BotError::MongoDbError
                }
            })
        })?;

        Ok(guild)
    }

    async fn get_or_create(&self, id: u64) -> Result<Guild, BotError> {
        let result = self.get(id).await;

        match result {
            Ok(v) => Ok(v),
            Err(e) => match e {
                BotError::TicketGuildNotFound => self.create(id).await,
                _ => Err(e),
            },
        }
    }

    async fn update_or_create(&self, id: u64, data: UpdateGuildData) -> Result<Guild, BotError> {
        let find = self
            .col
            .find_one(doc! {"_id": id as i64}, None)
            .await
            .or_else(|e| {
                log::error!(target: "mongo_errors", "{e}");
                Err(BotError::MongoDbError)
            })?;

        if let Some(guild) = find {
            self.update(id, data).await?;
            Ok(guild)
        } else {
            let guild = Guild {
                allow_multiple: data.allow_multiple.unwrap_or(false),
                enable_tickets: data.enable_tickets.unwrap_or(false),
                id: id as i64,
            };

            match self.col.insert_one(&guild, None).await {
                Ok(_) => Ok(guild),
                Err(e) => {
                    log::error!(target: "mongo_errors", "{e}");
                    Err(BotError::MongoDbError)
                }
            }
        }
    }
}
