
use std::io::{self, ErrorKind};
use std::env;
use std::os::unix::net::UnixDatagram;

pub fn notify(state: &str) -> io::Result<()> {
    let socket_path = match env::var_os("NOTIFY_SOCKET") {
        Some(path) => path,
        None => return Ok(()),
    };
    let sock = UnixDatagram::unbound()?;
    let len = sock.send_to(state.as_bytes(), socket_path)?;
    if len != state.len() {
        Err(io::Error::new(ErrorKind::WriteZero, "incomplete write"))
    } else {
        Ok(())
    }
}

pub fn notify_service_ready() {
    match notify("READY=1") {
        Ok(_) => (),
        Err(e) => {
            log::warn!("unable to notify host os about READY with {}", e);
            ()
        },
    }
}

pub fn notify_service_stopping() {
    match notify("STOPPING=1") {
        Ok(_) => (),
        Err(e) => {
            log::warn!("unable to notify host os about STOPPING with {}", e);
            ()
        },
    }
}
