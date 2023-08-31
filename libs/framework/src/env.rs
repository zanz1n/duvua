use std::{
    env, fmt,
    io::{Error, ErrorKind},
    str::FromStr,
};

#[derive(Debug, Clone)]
pub enum ProcessEnv {
    Development,
    Production,
    Testing,
    None,
}

impl FromStr for ProcessEnv {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "DEV" | "DEVELOPMENT" => Ok(Self::Development),
            "PROD" | "PRODUCTION" => Ok(Self::Production),
            "TEST" | "TESTING" => Ok(Self::Testing),
            _ => Err(Error::new(
                ErrorKind::InvalidData,
                "The value must be: DEVELOPMENT | DEV | PROD | PRODUCTION",
            )),
        }
    }
}

pub fn env_param<T>(key: &str, default: Option<T>) -> T
where
    T: FromStr + fmt::Debug,
    <T as FromStr>::Err: fmt::Debug,
{
    if let Ok(value) = env::var(key) {
        return value.parse::<T>().unwrap_or_else(|err| match default {
            Some(v) => {
                log::error!(
                    "Environment variable '{key}' must be valid, using default value '{v:?}'"
                );
                v
            }
            None => panic!("Environment variable '{key}' must be valid: {err:?}"),
        });
    } else {
        match default {
            Some(v) => v,
            None => panic!("Environment variable '{key}' is required"),
        }
    }
}
