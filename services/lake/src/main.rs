use std::os::unix::net::UnixDatagram;
use std::process::exit;
use std::sync::atomic::AtomicBool;
//use std::sync::Arc;
use std::{env, io};

use log::LevelFilter;
//use tokio::io::Error;
//use tokio::sync::mpsc;
//use zeromq::*;
use zmq;
use zmq_sys;
//use zmq_sys::libc::c_int;

use libc::c_int;

use crate::config::Configuration;
use crate::message::{msg_ptr, Message};
use crate::relay::{Context, Socket};

use crate::metrics::MetricCmdType::{EGRESS, INGRESS};
use crate::metrics::Metrics;
use crate::program::Program;

mod config;
mod message;
mod metrics;
mod program;
mod relay;

// #[tokio::main(flavor = "multi_thread")]
//#[tokio::main(flavor = "current_thread")]
fn main() -> Result<(), zmq::Error> {
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
    // INFO does not unbinds on drop

    let publisher = Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB as c_int)?;
    publisher.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1)?;
    publisher.set_option(zmq_sys::ZMQ_LINGER as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_SNDHWM as c_int, 0)?;
    publisher.bind(&format!("tcp://127.0.0.1:{}", config.pub_port))?;

    let metrics = Metrics::new(&config);

    //let (sub_results_sender, mut sub_results) = mpsc::channel::<ZmqMessage>(10_000_000);
    let metrics_sender_1 = metrics.sender.clone();
    //let metrics_sender_2 = metrics.sender.clone();

    // INFO does not unbinds on drop

    // FIXME inline
    loop {
        let mut msg = Message::new();
        let ptr = msg_ptr(&mut msg);
        if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as c_int) } == -1 {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        let _ = metrics_sender_1.send(INGRESS);
        //metrics.message_ingress(); // FIXME this seems to be slowing the code from 23s to 53s (1M -> 500k / sec)
        if unsafe {
            let data = zmq_sys::zmq_msg_data(ptr);
            let len = zmq_sys::zmq_msg_size(ptr) as usize;
            zmq_sys::zmq_send(publisher.sock, data, len, 0 as c_int)
        } == -1
        {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        let _ = metrics_sender_1.send(EGRESS);
        //metrics.message_egress(); // FIXME this seems to be slowing the code from 23s to 53s (1M -> 500k / sec)
    }

    Ok(())
    /*
    tokio::spawn(async move {
        loop {
            match sub_results.recv().await {
                Some(m) => match socket_pub.send(m).await {
                    Ok(_) => {
                        let _ = metrics_sender_1.send(EGRESS);
                    }
                    Err(e) => {
                        eprintln!("ZMQ error");
                        stopping();
                        exit(1);
                    }
                },
                None => {
                    eprintln!("Error processing queue");
                    stopping();
                    exit(1)
                }
            }
        }
    });

    loop {
        match socket_pull.recv().await {
            Ok(m) => {
                let _ = metrics_sender_2.send(INGRESS);
                match sub_results_sender.send(m).await {
                    Ok(_) => {}
                    Err(e) => {
                        eprintln!("Error sending to queue");
                        stopping();
                        exit(1);
                    }
                }
            }
            Err(_) => {
                eprintln!("ZMQ error");
                stopping();
                exit(0)
            }
        }
    }
    */
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
