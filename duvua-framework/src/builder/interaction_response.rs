use super::hashmap_to_json_map;
use crate::errors::BotError;
use serde_json::{Map, Value};
use serenity::{
    builder::CreateInteractionResponseData,
    http::Http,
    json::ToNumber,
    model::prelude::{
        application_command::ApplicationCommandInteraction,
        message_component::MessageComponentInteraction, AttachmentType, InteractionResponseType,
    },
};
use std::collections::HashMap;

#[derive(Clone, Debug)]
pub struct InteractionResponse<'a>(
    pub HashMap<&'static str, Value>,
    pub Vec<AttachmentType<'a>>,
);

impl<'a> InteractionResponse<'a> {
    #[inline]
    pub fn with_content<T: ToString>(content: T) -> Self {
        let mut data = Map::with_capacity(1);
        data.insert("content".to_owned(), Value::String(content.to_string()));

        let mut map = HashMap::<&'static str, Value>::with_capacity(2);
        map.insert("type", 4.to_number());
        map.insert("data", Value::Object(data));

        Self(map, Vec::with_capacity(0))
    }

    #[inline]
    pub fn set_kind(&mut self, kind: InteractionResponseType) -> &mut Self {
        self.0.insert("type", (kind as u8).to_number());
        self
    }

    #[inline]
    pub fn set_data(&mut self, data: &mut CreateInteractionResponseData<'a>) -> &mut Self {
        let data = data.to_owned();

        let map = hashmap_to_json_map(data.0);

        self.0.insert("data", Value::from(map));
        self.1 = data.1;
        self
    }

    pub async fn respond_message_component(
        &mut self,
        http: impl AsRef<Http>,
        data: &MessageComponentInteraction,
    ) -> Result<(), BotError> {
        self.to_owned().respond(http, data.id.0, &data.token).await
    }

    pub async fn respond_application_command(
        &mut self,
        http: impl AsRef<Http>,
        data: &ApplicationCommandInteraction,
    ) -> Result<(), BotError> {
        self.to_owned().respond(http, data.id.0, &data.token).await
    }

    pub async fn respond(
        self,
        http: impl AsRef<Http>,
        interaction_id: u64,
        interaction_token: &str,
    ) -> Result<(), BotError> {
        let map = hashmap_to_json_map(self.0);
        let map = Value::from(map);

        if self.1.is_empty() {
            http.as_ref()
                .create_interaction_response(interaction_id, interaction_token, &map)
                .await
        } else {
            http.as_ref()
                .create_interaction_response_with_files(
                    interaction_id,
                    interaction_token,
                    &map,
                    self.1,
                )
                .await
        }
        .or_else(|e| Err(BotError::Serenity(e)))
    }
}

impl<'a> Default for InteractionResponse<'a> {
    fn default() -> InteractionResponse<'a> {
        let mut map = HashMap::new();
        map.insert("type", 4.to_number());

        Self(map, Vec::new())
    }
}
