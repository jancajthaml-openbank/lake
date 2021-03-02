use std::env;
use std::io;
use std::os::unix::net::UnixDatagram;

pub fn notify(state: &str) -> io::Result<()> {
    let socket_path = match env::var_os("NOTIFY_SOCKET") {
        Some(path) => path,
        None => return Ok(()),
    };
    let sock = UnixDatagram::unbound()?;
    let len = sock.send_to(state.as_bytes(), socket_path)?;
    if len == state.len() {
        Ok(())
    } else {
        Err(io::Error::new(io::ErrorKind::WriteZero, "incomplete write"))
    }
}

pub fn notify_service_ready() {
    match notify("READY=1") {
        Ok(_) => (),
        Err(e) => {
            log::warn!("unable to notify host os about READY with {}", e);
        }
    }
}

pub fn notify_service_stopping() {
    match notify("STOPPING=1") {
        Ok(_) => (),
        Err(e) => {
            log::warn!("unable to notify host os about STOPPING with {}", e);
        }
    }
}
