use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct KissCacheData {
    pub user_id: u64,
    pub target_id: u64,
    pub interaction_token: String,
}

impl KissCacheData {
    pub fn new(user_id: u64, target_id: u64, interaction_id: String) -> Self {
        Self {
            user_id,
            target_id,
            interaction_token: interaction_id,
        }
    }
}
