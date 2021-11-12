use std::os::unix::net::UnixDatagram;
use std::{env, io};

use zmq_sys;

use libc::c_int;

use crate::config::Configuration;
use crate::message::{msg_ptr, Message};
use crate::socket::{Context, Socket};

use crate::metrics::MetricCmdType::{EGRESS, INGRESS};
use crate::metrics::Metrics;

//use signal_hook::consts::{SIGQUIT, TERM_SIGNALS};
//use signal_hook::iterator::Signals;

mod config;
mod error;
mod logging;
mod message;
mod metrics;
mod socket;

fn main() -> Result<(), String> {
    let config = Configuration::load();

    let _ = logging::setup_logging(&config);

    ready();

    let ctx = Context::new();

    let puller = match setup_pull_socket(&ctx, &config) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PULL socket".to_owned()),
    };

    let publisher = match setup_pub_socket(&ctx, &config) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PUB socket".to_owned()),
    };

    let metrics = match Metrics::new(&config) {
        Ok(instance) => instance,
        Err(_) => return Err("unable to instantiate metrics".to_owned()),
    };
    let metrics_sender_1 = metrics.sender.clone();

    // info create program here
    ready();

    // FIXME need somehow kill the loop below
    //let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
    //let _ = sigs.wait();

    println!("entering loop");

    loop {
        let mut msg = Message::new();
        let ptr = msg_ptr(&mut msg);
        if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as c_int) } == -1 {
            // FIXME stopping as a drop of program
            stopping();
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }).to_string());
        };
        let _ = metrics_sender_1.send(INGRESS);
        if unsafe {
            let data = zmq_sys::zmq_msg_data(ptr);
            let len = zmq_sys::zmq_msg_size(ptr) as usize;
            zmq_sys::zmq_send(publisher.sock, data, len, 0 as c_int)
        } == -1
        {
            // FIXME stopping as a drop of program
            stopping();
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }).to_string());
        };
        let _ = metrics_sender_1.send(EGRESS);
    }
}

fn setup_pull_socket(ctx: &Context, config: &Configuration) -> Result<Socket, String> {
    let puller = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PULL as c_int) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PULL socket".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_LINGER as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_RCVHWM as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_RCVHWM option to 0".to_owned()),
    };
    match puller.bind(&format!("tcp://127.0.0.1:{}", config.pull_port)) {
        Ok(_) => Ok(puller),
        Err(_) => return Err("unable to bind PULL socket".to_owned()),
    }
}

fn setup_pub_socket(ctx: &Context, config: &Configuration) -> Result<Socket, String> {
    let publisher = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB as c_int) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PUB socket".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_LINGER as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_SNDHWM as c_int, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_SNDHWM option to 0".to_owned()),
    };
    match publisher.bind(&format!("tcp://127.0.0.1:{}", config.pub_port)) {
        Ok(_) => Ok(publisher),
        Err(_) => return Err("unable to bind PUB socket".to_owned()),
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
