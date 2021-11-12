use crate::config::Configuration;

use log::LevelFilter;
use simple_logger::SimpleLogger;

pub fn setup_logging(config: &Configuration) -> Result<(), String> {
    match SimpleLogger::new().init() {
        Ok(_) => {}
        Err(_) => return Err("unable to initialize logger".to_owned()),
    };

    log::set_max_level(LevelFilter::Info);

    let level = match &*config.log_level {
        "DEBUG" => LevelFilter::Debug,
        "INFO" => LevelFilter::Info,
        "WARN" => LevelFilter::Warn,
        "ERROR" => LevelFilter::Error,
        _ => {
            log::warn!("Invalid log level {}, using level INFO", config.log_level);
            LevelFilter::Info
        }
    };

    log::info!("Log level set to {}", level.as_str());
    log::set_max_level(level);

    Ok(())
}
