use crate::{
    test_utils::{check_integration_test_environment, prepare_environment},
    welcome::{Welcome, WelcomeRepository, WelcomeService, WelcomeType},
};
use chrono::{NaiveDateTime, Utc};
use std::time::Duration;

async fn create_repo() -> WelcomeService {
    let pool = match prepare_environment().await {
        Ok(v) => v,
        Err(e) => {
            panic!("Failed to load test environment: {e}");
        }
    };

    WelcomeService::new(pool)
}

async fn assert_creation(service: &WelcomeService, returned: Welcome, start: NaiveDateTime) {
    let time_dif = returned.created_at.timestamp_millis() - start.timestamp_millis();

    if 10 * 1000 < time_dif {
        service.delete_by_id(returned.id).await.unwrap();
        panic!("Discrepant database/computer timestamps diff! Maybe timezone issues?")
    }

    let find_welcome = service.get_by_id(returned.id).await.unwrap().unwrap();

    assert_eq!(returned.id, find_welcome.id);
    assert_eq!(returned.created_at, find_welcome.created_at);
    assert_eq!(returned.updated_at, find_welcome.updated_at);
    assert_eq!(returned.enabled, find_welcome.enabled);
    assert_eq!(returned.channel_id, find_welcome.channel_id);
    assert_eq!(returned.message, find_welcome.message);
    assert_eq!(returned.kind, find_welcome.kind);
}

#[tokio::test]
async fn test_create_default() {
    const FAKE_ID: i64 = 867141919517976080;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;
    let start = Utc::now().naive_utc();
    let welcome = service.create_default(FAKE_ID).await.unwrap();

    let welcome_id = welcome.id;

    assert_creation(&service, welcome, start).await;

    assert!(service.delete_by_id(welcome_id).await.unwrap())
}

#[tokio::test]
async fn test_create() {
    const FAKE_ID: i64 = 967741919217976080;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;
    let start = Utc::now().naive_utc();

    let welcome = Welcome {
        id: FAKE_ID,
        created_at: start,
        updated_at: start,
        enabled: true,
        channel_id: Some(967141216511977080),
        message: "Fake Message".to_owned(),
        kind: WelcomeType::Image,
    };

    let welcome_id = welcome.id;

    let created = service.create(welcome.clone()).await.unwrap();

    assert_creation(&service, created, start).await;

    assert!(service.delete_by_id(welcome_id).await.unwrap());
}

#[tokio::test]
async fn test_exists() {
    const FAKE_ID: i64 = 367741919517976080;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;
    let welcome = service.create_default(FAKE_ID).await.unwrap();

    assert_eq!(FAKE_ID, welcome.id);

    let exists = service.exists(welcome.id).await.unwrap();

    assert!(exists);

    assert!(service.delete_by_id(welcome.id).await.unwrap());
}

#[tokio::test]
async fn test_update_enabled() {
    const FAKE_ID: i64 = 267741219417956680;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;

    let welcome = service.create_default(FAKE_ID).await.unwrap();

    tokio::time::sleep(Duration::from_millis(100)).await;

    service
        .update_enabled(welcome.id, !welcome.enabled)
        .await
        .unwrap();

    let find = service.get_by_id(welcome.id).await.unwrap().unwrap();

    assert_eq!(!find.enabled, welcome.enabled);
    assert_ne!(find.updated_at, welcome.updated_at);

    assert!(service.delete_by_id(welcome.id).await.unwrap());
}

#[tokio::test]
async fn test_update_channel_id() {
    const FAKE_CHANNEL_ID: i64 = 1240584214774735599;
    const FAKE_ID: i64 = 961741219412956680;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;
    let welcome = service.create_default(FAKE_ID).await.unwrap();

    tokio::time::sleep(Duration::from_millis(100)).await;

    service
        .update_channel_id(welcome.id, Some(FAKE_CHANNEL_ID))
        .await
        .unwrap();

    let find = service.get_by_id(welcome.id).await.unwrap().unwrap();

    assert_eq!(find.enabled, welcome.enabled);
    assert_eq!(Some(FAKE_CHANNEL_ID), find.channel_id);
    assert_ne!(find.updated_at, welcome.updated_at);

    assert!(service.delete_by_id(welcome.id).await.unwrap());
}

#[tokio::test]
async fn test_update_message() {
    const FAKE_MESSAGE: &str = "Some new message";
    const FAKE_ID: i64 = 911336173564129627;
    if !check_integration_test_environment() {
        println!("Skiping integration tests");
        return;
    }

    let service = create_repo().await;
    let welcome = service.create_default(FAKE_ID).await.unwrap();

    tokio::time::sleep(Duration::from_millis(100)).await;

    service
        .update_message(welcome.id, FAKE_MESSAGE.to_owned(), WelcomeType::Image)
        .await
        .unwrap();

    let find = service.get_by_id(welcome.id).await.unwrap().unwrap();

    assert_eq!(find.enabled, welcome.enabled);
    assert_eq!(FAKE_MESSAGE.to_owned(), find.message);
    assert_eq!(WelcomeType::Image, find.kind);
    assert_ne!(find.updated_at, welcome.updated_at);

    assert!(service.delete_by_id(welcome.id).await.unwrap());
}
