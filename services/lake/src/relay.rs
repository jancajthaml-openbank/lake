use log;
use signal_hook::consts::SIGQUIT;
use signal_hook::low_level;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;
use zmq_sys;

use crate::config::Configuration;
use crate::error;
use crate::message::{msg_ptr, Message};
use crate::metrics::Metrics;
use crate::socket::{Context, Socket};

/// zmq relay subroutine
pub struct Relay {
    child_thread: Option<thread::JoinHandle<()>>,
    kill_port: i32,
}

impl Drop for Relay {
    fn drop(&mut self) {
        log::info!("Relay stopping");
        let ctx = Context::new();
        match Socket::new(ctx.underlying, zmq_sys::ZMQ_PUSH as i32) {
            Ok(pusher) => {
                log::debug!("Relay unblocking child thread");
                match pusher.connect(&format!("tcp://127.0.0.1:{}", self.kill_port)) {
                    Ok(_) => {
                        let mut msg = Message::new();
                        let ptr = msg_ptr(&mut msg);
                        unsafe { zmq_sys::zmq_msg_send(ptr, pusher.sock, 0 as i32) };
                    }
                    Err(_) => {}
                };
                drop(pusher);
            }
            Err(_) => {}
        };
        log::debug!("Relay waiting for child thread to terminate");
        let _ = self.child_thread.take().unwrap().join();
        log::debug!("Relay child thread exited");
        log::info!("Relay stopped");
    }
}

impl Relay {
    /// creates new relay fascade
    #[must_use]
    pub fn new(
        config: &Configuration,
        prog_running: Arc<AtomicBool>,
        metrics: Arc<Metrics>,
    ) -> Arc<Relay> {
        log::info!("Relay starting");

        let pull_port = config.pull_port;
        let pub_port = config.pub_port;

        let child_thread = thread::spawn(move || {
            let ctx = Context::new();
            let _ = ctx.set_io_threads(num_cpus::get());

            let puller = match setup_pull_socket(&ctx, pull_port) {
                Ok(sock) => Some(sock),
                Err(err) => {
                    log::error!("unable to initialize PULL socket {}", err);
                    None
                }
            };

            let publisher = match setup_pub_socket(&ctx, pub_port) {
                Ok(sock) => Some(sock),
                Err(err) => {
                    log::error!("unable to initialize PUB socket {}", err);
                    None
                }
            };

            match (puller, publisher) {
                (Some(puller), Some(publisher)) => {
                    log::info!("Relay started");
                    loop {
                        let mut msg = Message::new();
                        let ptr = msg_ptr(&mut msg);
                        if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as i32) } == -1 {
                            log::error!(
                                "{}",
                                error::Error::from_raw(unsafe { zmq_sys::zmq_errno() })
                            );
                            break;
                        };
                        if !prog_running.load(Ordering::Relaxed) {
                            break;
                        }
                        metrics.message_ingress();
                        if unsafe { zmq_sys::zmq_msg_send(ptr, publisher.sock, 0 as i32) } == -1 {
                            log::error!(
                                "{}",
                                error::Error::from_raw(unsafe { zmq_sys::zmq_errno() })
                            );
                            break;
                        };
                        metrics.message_egress();
                    }
                }
                _ => {}
            }

            let _ = low_level::raise(SIGQUIT);
        });

        Arc::new(Relay {
            kill_port: config.pull_port,
            child_thread: Some(child_thread),
        })
    }
}

fn setup_pull_socket(ctx: &Context, port: i32) -> Result<Socket, String> {
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
    match puller.bind(&format!("tcp://127.0.0.1:{}", port)) {
        Ok(_) => Ok(puller),
        Err(_) => return Err("unable to bind PULL socket".to_owned()),
    }
}

fn setup_pub_socket(ctx: &Context, port: i32) -> Result<Socket, String> {
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
    match publisher.bind(&format!("tcp://127.0.0.1:{}", port)) {
        Ok(_) => Ok(publisher),
        Err(_) => return Err("unable to bind PUB socket".to_owned()),
    }
}
