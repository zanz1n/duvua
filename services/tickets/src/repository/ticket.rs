use async_trait::async_trait;
use chrono::{DateTime, Utc};
use duvua_framework::errors::BotError;
use mongodb::{
    bson::{doc, oid::ObjectId, Document},
    Collection,
};
use serde::{Deserialize, Serialize};
use std::str::FromStr;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Ticket {
    #[serde(rename = "_id")]
    pub id: ObjectId,
    #[serde(with = "mongodb::bson::serde_helpers::chrono_datetime_as_bson_datetime")]
    pub created_at: DateTime<Utc>,
    pub channel_id: i64,
    pub user_id: i64,
    pub guild_id: i64,
}

impl Ticket {
    #[inline]
    pub fn from_data(data: CreateTicketData) -> Self {
        Self {
            id: ObjectId::new(),
            created_at: Utc::now(),
            channel_id: data.channel_id,
            guild_id: data.guild_id,
            user_id: data.user_id,
        }
    }
}

#[derive(Debug, Clone)]
pub struct CreateTicketData {
    pub id: ObjectId,
    pub channel_id: i64,
    pub user_id: i64,
    pub guild_id: i64,
}

impl CreateTicketData {
    #[inline]
    pub fn from_snowflakes(channel_id: u64, user_id: u64, guild_id: u64) -> Self {
        Self {
            id: ObjectId::new(),
            channel_id: channel_id as i64,
            user_id: user_id as i64,
            guild_id: guild_id as i64,
        }
    }
}

#[async_trait]
pub trait TicketRepository: Sync + Send {
    async fn get(&self, id: String) -> Result<Ticket, BotError>;
    async fn get_by_member(&self, guild_id: u64, user_id: u64) -> Result<Ticket, BotError>;
    async fn create(&self, data: CreateTicketData) -> Result<Ticket, BotError>;
    async fn delete(&self, id: String) -> Result<(), BotError>;
}

fn parse_object_id(id: &str) -> Result<ObjectId, BotError> {
    ObjectId::from_str(id).or_else(|_| Err(BotError::InvalidMongoDbObjectId))
}

pub struct TicketService {
    col: Collection<Ticket>,
}

impl TicketService {
    pub fn new(col: Collection<Ticket>) -> Self {
        Self { col }
    }

    pub async fn fetch(&self, filter: Document) -> Result<Ticket, BotError> {
        self.col
            .find_one(filter, None)
            .await
            .or_else(|e| {
                log::error!(target: "mongo_errors", "{e}");
                Err(BotError::MongoDbError)
            })?
            .ok_or(BotError::TicketNotFound)
    }
}

#[async_trait]
impl TicketRepository for TicketService {
    async fn get(&self, id: String) -> Result<Ticket, BotError> {
        let object_id = parse_object_id(id.as_str())?;

        self.fetch(doc! {"_id": object_id}).await
    }

    async fn get_by_member(&self, guild_id: u64, user_id: u64) -> Result<Ticket, BotError> {
        self.col
            .find_one(
                doc! {"guild_id": guild_id as i64, "user_id": user_id as i64},
                None,
            )
            .await
            .or_else(|_| Err(BotError::MongoDbError))?
            .ok_or(BotError::TicketNotFound)
    }

    async fn create(&self, data: CreateTicketData) -> Result<Ticket, BotError> {
        let ticket = Ticket::from_data(data);

        self.col
            .insert_one(&ticket, None)
            .await
            .or_else(|_| Err(BotError::MongoDbError))?;

        Ok(ticket)
    }

    async fn delete(&self, id: String) -> Result<(), BotError> {
        let object_id = parse_object_id(id.as_str())?;

        let res = self
            .col
            .delete_one(doc! {"_id": object_id}, None)
            .await
            .or_else(|e| {
                log::error!(target: "mongo_errors", "{e}");
                Err(BotError::MongoDbError)
            })?;

        if res.deleted_count == 0 {
            return Err(BotError::TicketNotFound);
        }

        Ok(())
    }
}
