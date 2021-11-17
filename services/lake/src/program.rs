//use log::LevelFilter;
//use simple_logger::SimpleLogger;
use std::os::unix::net::UnixDatagram;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::{env, io};

use crate::config::Configuration;
use crate::logger::Logger;

pub struct Program {
    pub running: Arc<AtomicBool>,
}

impl Program {
    pub fn new(config: &Configuration) -> Program {
        let prog = Program {
            running: Arc::new(AtomicBool::new(false)),
        };
        let _ = Logger::new().init(&*config.log_level);
        log::info!("Program starting");
        if let Err(e) = notify("READY=1") {
            log::warn!("unable to notify host os about READY with {}", e);
        };
        prog.running.store(true, Ordering::Relaxed);
        prog
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

#[cfg(target_family = "unix")]
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

#[cfg(not(target_family = "unix"))]
fn notify(msg: &msg) -> io::Result<()> {
    Ok(())
}
