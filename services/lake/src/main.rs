use log;
use signal_hook::consts::{SIGQUIT, TERM_SIGNALS};
use signal_hook::iterator::Signals;
use signal_hook::low_level;
use std::thread;
use zmq_sys;

use crate::config::Configuration;
use crate::message::{msg_ptr, Message};
use crate::metrics::Metrics;
use crate::program::Program;
use crate::socket::{Context, Socket};

mod config;
mod error;
mod message;
mod metrics;
mod program;
mod socket;

fn main() -> Result<(), String> {
    println!("main before");
    let x = run();
    println!("run after");
    x
}

fn run() -> Result<(), String> {
    let config = Configuration::load();

    let prog = Program::new(&config); // FIXME program should hold atomic representing app running / dead

    let metrics = match Metrics::new(&config) {
        Ok(instance) => instance,
        Err(_) => return Err("unable to instantiate metrics".to_owned()),
    };

    // TODO relay structure / fascade that is droppable

    //let metrics_1 = metrics.clone();

    thread::spawn(move || {
        let ctx = Context::new();
        let _ = ctx.set_io_threads(num_cpus::get());

        let puller = match setup_pull_socket(&ctx, &config) {
            Ok(sock) => Some(sock),
            Err(err) => {
                log::error!("unable to initialize PULL socket {}", err);
                None
            }
        };

        let publisher = match setup_pub_socket(&ctx, &config) {
            Ok(sock) => Some(sock),
            Err(err) => {
                log::error!("unable to initialize PUB socket {}", err);
                None
            }
        };

        match (puller, publisher) {
            (Some(puller), Some(publisher)) => loop {
                // FIXME check if application is dead
                let mut msg = Message::new();
                let ptr = msg_ptr(&mut msg);
                if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as i32) } == -1 {
                    log::error!(
                        "{}",
                        error::Error::from_raw(unsafe { zmq_sys::zmq_errno() })
                    );
                    break;
                };
                metrics.message_ingress();
                if unsafe { zmq_sys::zmq_msg_send(ptr, publisher.sock, 0 as i32) } == -1 {
                    log::error!(
                        "{}",
                        error::Error::from_raw(unsafe { zmq_sys::zmq_errno() })
                    );
                    break;
                };
                metrics.message_egress();
            },
            _ => {}
        }

        let _ = low_level::raise(SIGQUIT);
    });

    let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
    let _ = sigs.wait();

    // TODO drop relay
    // TODO drop metrics

    drop(prog);

    Ok(())
}

fn setup_pull_socket(ctx: &Context, config: &Configuration) -> Result<Socket, String> {
    let puller = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PULL as i32) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PULL socket".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_CONFLATE as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_IMMEDIATE as i32, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_LINGER as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_RCVHWM as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_RCVHWM option to 0".to_owned()),
    };
    match puller.bind(&format!("tcp://127.0.0.1:{}", config.pull_port)) {
        Ok(_) => Ok(puller),
        Err(_) => return Err("unable to bind PULL socket".to_owned()),
    }
}

fn setup_pub_socket(ctx: &Context, config: &Configuration) -> Result<Socket, String> {
    let publisher = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB as i32) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PUB socket".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_CONFLATE as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_IMMEDIATE as i32, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_LINGER as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_SNDHWM as i32, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_SNDHWM option to 0".to_owned()),
    };
    match publisher.bind(&format!("tcp://127.0.0.1:{}", config.pub_port)) {
        Ok(_) => Ok(publisher),
        Err(_) => return Err("unable to bind PUB socket".to_owned()),
    }
}
