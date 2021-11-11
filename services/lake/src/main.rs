use std::os::unix::net::UnixDatagram;
use std::process::exit;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;
use std::{env, io};

use log::LevelFilter;
use tokio::io::Error;
use tokio::sync::mpsc;
use zeromq::*;

use crate::config::Configuration;
use crate::metrics::Metrics;
use crate::program::Program;

mod config;
mod metrics;
mod program;

// #[tokio::main(flavor = "multi_thread")]
#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), Error> {
    let config = Configuration::load();
    let program = Program::new();
    let _ = program.setup(); // for logging now only

    ready();

    let mut socket_pull = zeromq::PullSocket::new();
    socket_pull
        .bind(&format!("tcp://127.0.0.1:{}", config.pull_port))
        .await
        .expect("Failed to bind PULL socket");

    let mut socket_pub = zeromq::PubSocket::new();
    socket_pub
        .bind(&format!("tcp://127.0.0.1:{}", config.pub_port))
        .await
        .expect("Failed to bind PUB socket");

    let metrics = Metrics::new(&config);

    let (sub_results_sender, mut sub_results) = mpsc::channel::<ZmqMessage>(10);

    tokio::spawn(async move {
        loop {
            match sub_results.recv().await {
                Some(m) => {
                    let _ = socket_pub.send(m);
                    // metrics.message_egress();  TODO enable here
                }
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
                metrics.message_ingress();
                match sub_results_sender.send(m).await {
                    Ok(_) => {
                        metrics.message_egress(); // TODO move up to send
                    }
                    Err(e) => {
                        eprintln!("Error sending to queue");
                        stopping();
                        exit(1);
                    }
                }
            }
            Err(_) => {
                eprintln!("ZMQ error");
                stopping()
            }
        }
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
