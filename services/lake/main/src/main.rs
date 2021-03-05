use boot::Program;
use lazy_static::lazy_static;
use std::env;
use std::io;
use std::os::unix::net::UnixDatagram;

fn main() {
    lazy_static! {
        static ref PROGRAM: Program = Program::new();
    }
    if let Err(e) = PROGRAM.setup() {
        panic!(e);
    };
    ready();
    if let Err(e) = PROGRAM.start() {
        stopping();
        panic!(e);
    };
    PROGRAM.stop();
    stopping();
}

fn ready() {
    if let Err(e) = notify("READY=1") {
        println!("unable to notify host os about READY with {}", e);
    }
}

fn stopping() {
    if let Err(e) = notify("STOPPING=1") {
        println!("unable to notify host os about STOPPING with {}", e)
    }
}

fn notify(state: &str) -> io::Result<()> {
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
