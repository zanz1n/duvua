use crate::{
    guild::{GuildRepository, GuildService},
    test_utils::{check_integration_test_environment, prepare_environment},
};

async fn create_repo() -> GuildService {
    let pool = match prepare_environment().await {
        Ok(v) => v,
        Err(e) => {
            panic!("Failed to load test environment: {e}");
        }
    };

    GuildService::new(pool)
}

#[tokio::test]
async fn test_exists() {
    const FAKE_ID: i64 = 867141919517976080;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;

    let exists = service.exists(FAKE_ID).await.unwrap();
    assert!(!exists);

    _ = service.create_default(FAKE_ID).await.unwrap();
    let exists = service.exists(FAKE_ID).await.unwrap();

    assert!(exists);

    assert!(service.delete_by_id(FAKE_ID).await.unwrap());
}
