use log::LevelFilter;
use simple_logger::SimpleLogger;
use std::os::unix::net::UnixDatagram;
use std::{env, io};

use crate::config::Configuration;

pub struct Program {}

impl Program {
    pub fn new(config: &Configuration) -> Program {
        let prog = Program {};
        let _ = prog.setup_logging(config);
        log::info!("Program starting");
        if let Err(e) = notify("READY=1") {
            log::warn!("unable to notify host os about READY with {}", e);
        };
        prog
    }

    fn setup_logging(&self, config: &Configuration) -> Result<(), String> {
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
}

impl Drop for Program {
    fn drop(&mut self) {
        log::info!("Program stopping");
        if let Err(e) = notify("STOPPING=1") {
            log::warn!("unable to notify host os about STOPPING with {}", e)
        }
    }
}

/// sends msg to `NOTIFY_SOCKET` via udp
fn notify(msg: &str) -> io::Result<()> {
    let socket_path = match env::var_os("NOTIFY_SOCKET") {
        Some(path) => path,
        None => {
            return Ok(());
        }
    };
    let sock = UnixDatagram::unbound()?;
    let len = sock.send_to(msg.as_bytes(), socket_path)?;
    if len == msg.len() {
        Ok(())
    } else {
        Err(io::Error::new(io::ErrorKind::WriteZero, "incomplete write"))
    }
}
