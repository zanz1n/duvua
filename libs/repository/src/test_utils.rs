use sqlx::migrate::Migrator;
use std::env;

static MIGRATIONS: Migrator = sqlx::migrate!("../../migrations");

pub fn check_integration_test_environment() -> bool {
    _ = dotenvy::dotenv();

    if env::var("DATABASE_URL").unwrap_or("".to_owned()).len() < 2 {
        return false;
    };

    let env = env::var("DUVUA_TESTING_POSTGRES_INTEGRATION").unwrap_or("".to_owned());

    env == "true" || env == "TRUE" || env == "1"
}

pub async fn prepare_environment() -> Result<sqlx::PgPool, Box<dyn std::error::Error>> {
    _ = dotenvy::dotenv();

    let url = env::var("DATABASE_URL").unwrap();

    let db_pool = sqlx::PgPool::connect(&url).await?;
    MIGRATIONS.run(&db_pool).await?;

    Ok(db_pool)
}
