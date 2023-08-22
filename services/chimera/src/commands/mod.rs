pub mod avatar;
pub mod clone;
pub mod facts;
pub mod kiss;
pub mod ping;

use base64::{engine::general_purpose::STANDARD, Engine};
use duvua_framework::errors::BotError;

pub(crate) async fn get_base64_image_data(url: &str) -> Result<String, BotError> {
    log::info!("{url}");
    let res = reqwest::get(url).await.or_else(|e| {
        log::error!("Failed to fetch discord cdn: {e}");
        Err(BotError::UserAvatarFetchFailed)
    })?;

    let buf = res.bytes().await.or_else(|e| {
        log::error!("Failed read discord cdn response body: {e}");
        Err(BotError::UserAvatarFetchFailed)
    })?;

    Ok(STANDARD.encode(buf))
}
