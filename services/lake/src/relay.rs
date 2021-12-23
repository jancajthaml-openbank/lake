use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;

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
        if let Ok(pusher) = Socket::new(ctx.underlying, zmq_sys::ZMQ_PUSH) {
            log::debug!("Relay unblocking child thread");
            if pusher
                .connect(&format!("tcp://127.0.0.1:{}", self.kill_port))
                .is_ok()
            {
                let mut msg = Message::new();
                let ptr = msg_ptr(&mut msg);
                unsafe { zmq_sys::zmq_msg_send(ptr, pusher.sock, 0_i32) };
                drop(pusher);
            };
        };
        log::debug!("Relay waiting for child thread to terminate");
        let _res = self.child_thread.take().unwrap().join();
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
        let ctx = Context::new();

        let child_thread = thread::spawn(move || {
            let _ = ctx.set_io_threads(1);

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

            if let (Some(puller), Some(publisher)) = (puller, publisher) {
                log::info!("Relay started");
                loop {
                    let mut msg = Message::new();
                    let ptr = msg_ptr(&mut msg);
                    if unsafe {
                        zmq_sys::zmq_msg_recv(ptr, puller.sock, 0_i32) == -1
                            || zmq_sys::zmq_msg_send(ptr, publisher.sock, 0_i32) == -1
                    } {
                        log::error!(
                            "{}",
                            error::Error::from_raw(unsafe { zmq_sys::zmq_errno() })
                        );
                        break;
                    };
                    if !prog_running.load(Ordering::Relaxed) {
                        break
                    }
                    metrics.relayed();
                }
                drop(puller);
                drop(publisher);
            }

            drop(ctx);
            unsafe { libc::raise(libc::SIGTERM) };
        });

        Arc::new(Relay {
            kill_port: config.pull_port,
            child_thread: Some(child_thread),
        })
    }
}

fn setup_pull_socket(ctx: &Context, port: i32) -> Result<Socket, String> {
    let puller = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PULL) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PULL socket".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_CONFLATE, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_IMMEDIATE, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_LINGER, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match puller.set_option(zmq_sys::ZMQ_RCVHWM, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PULL socket ZMQ_RCVHWM option to 0".to_owned()),
    };
    match puller.bind(&format!("tcp://0.0.0.0:{}", port)) {
        Ok(_) => {
            log::info!("Started PULL socket on 0.0.0.0:{}", port);
            Ok(puller)
        },
        Err(_) => Err("unable to bind PULL socket".to_owned()),
    }
}

fn setup_pub_socket(ctx: &Context, port: i32) -> Result<Socket, String> {
    let publisher = match Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB) {
        Ok(sock) => sock,
        Err(_) => return Err("unable to initialize PUB socket".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_CONFLATE, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_CONFLATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_IMMEDIATE, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 1".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_LINGER, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_IMMEDIATE option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_SNDHWM, 0) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_SNDHWM option to 0".to_owned()),
    };
    match publisher.set_option(zmq_sys::ZMQ_XPUB_NODROP, 1) {
        Ok(_) => {}
        Err(_) => return Err("unable to set PUB socket ZMQ_XPUB_NODROP option to 1".to_owned()),
    };
    match publisher.bind(&format!("tcp://0.0.0.0:{}", port)) {
        Ok(_) => {
            log::info!("Started PUB socket on 0.0.0.0:{}", port);
            Ok(publisher)
        },
        Err(_) => Err("unable to bind PUB socket".to_owned()),
    }
}
