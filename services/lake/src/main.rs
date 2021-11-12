use std::os::unix::net::UnixDatagram;
use std::{env, io};

use zmq_sys;

use libc::c_int;

use crate::config::Configuration;
use crate::message::{msg_ptr, Message};
use crate::socket::{Context, Socket};

//use crate::errors;
use crate::metrics::MetricCmdType::{EGRESS, INGRESS};
use crate::metrics::Metrics;
use crate::program::Program;

mod config;
mod error;
mod message;
mod metrics;
mod program;
mod socket;

fn main() -> Result<(), error::Error> {
    let config = Configuration::load();
    let program = Program::new();

    let _ = program.setup(); // for logging now only

    ready();

    let ctx = Context::new();

    let puller = Socket::new(ctx.underlying, zmq_sys::ZMQ_PULL as c_int)?;
    puller.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0)?;
    puller.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1)?;
    puller.set_option(zmq_sys::ZMQ_LINGER as c_int, 0)?;
    puller.set_option(zmq_sys::ZMQ_RCVHWM as c_int, 0)?;
    puller.bind(&format!("tcp://127.0.0.1:{}", config.pull_port))?;

    let publisher = Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB as c_int)?;
    publisher.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1)?;
    publisher.set_option(zmq_sys::ZMQ_LINGER as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_SNDHWM as c_int, 0)?;
    publisher.bind(&format!("tcp://127.0.0.1:{}", config.pub_port))?;

    let metrics = Metrics::new(&config);

    let metrics_sender_1 = metrics.sender.clone();

    loop {
        let mut msg = Message::new();
        let ptr = msg_ptr(&mut msg);
        if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as c_int) } == -1 {
            stopping();
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        let _ = metrics_sender_1.send(INGRESS);
        if unsafe {
            let data = zmq_sys::zmq_msg_data(ptr);
            let len = zmq_sys::zmq_msg_size(ptr) as usize;
            zmq_sys::zmq_send(publisher.sock, data, len, 0 as c_int)
        } == -1
        {
            stopping();
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        let _ = metrics_sender_1.send(EGRESS);
    }
}

/// tries to notify host os that service is ready
fn ready() {
    if let Err(e) = notify("READY=1") {
        println!("unable to notify host os about READY with {}", e);
    }
}

/// tries to notify host os that service is stopping
fn stopping() {
    if let Err(e) = notify("STOPPING=1") {
        println!("unable to notify host os about STOPPING with {}", e)
    }
}

/// sends msg to `NOTIFY_SOCKET` via udp
fn notify(msg: &str) -> io::Result<()> {
    let socket_path = match env::var_os("NOTIFY_SOCKET") {
        Some(path) => path,
        None => return Ok(()),
    };
    let sock = UnixDatagram::unbound()?;
    let len = sock.send_to(msg.as_bytes(), socket_path)?;
    if len == msg.len() {
        Ok(())
    } else {
        Err(io::Error::new(io::ErrorKind::WriteZero, "incomplete write"))
    }
}
