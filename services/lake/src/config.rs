use std::env;

pub struct Configuration {
    /// port to bind ZMQ PULL
    pub pull_port: i32,
    /// port to bind ZMQ PUB
    pub pub_port: i32,
    /// ignorecase log level
    pub log_level: String,
    /// statsd daemon endpoint
    pub statsd_endpoint: String,
}

impl Configuration {
    /// populates configuration from env
    #[must_use]
    pub fn load() -> Configuration {
        Configuration {
            pull_port: env_i32("LAKE_PORT_PULL", 5562),
            pub_port: env_i32("LAKE_PORT_PUB", 5561),
            log_level: env_string("LAKE_LOG_LEVEL", "INFO").to_uppercase(),
            statsd_endpoint: env_string("LAKE_STATSD_ENDPOINT", "127.0.0.1:8125"),
        }
    }
}

/// fetches optional value from env
fn env_get(key: &str) -> Option<String> {
    match env::var_os(key) {
        Some(val) => val.to_str().map(std::string::ToString::to_string),
        None => None,
    }
}

/// fetches str env value if present or yields fallback if missing
fn env_string(key: &str, fallback: &str) -> String {
    match env_get(key) {
        Some(val) => val,
        None => fallback.to_string(),
    }
}

/// fetches i32 env value if present or yields fallback if missing or invalid
fn env_i32(key: &str, fallback: i32) -> i32 {
    match env_get(key) {
        Some(untyped) => match untyped.parse::<i32>() {
            Ok(val) => val,
            Err(_e) => fallback,
        },
        None => fallback,
    }
}
