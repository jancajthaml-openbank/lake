use std::os::unix::net::UnixDatagram;
use std::process::exit;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;
use std::{env, io};

use tokio::io::Error;
use tokio::sync::mpsc;
use zeromq::*;

use crate::config::Configuration;
use crate::metrics::Metrics;

mod config;
mod metrics;

//#[tokio::main(flavor = "multi_thread")]
#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), Error> {
    let config = Configuration::load();

    // STUB for setup logging and bbtests
    println!("Log level set to {}", config.log_level);

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

    // FIXME better with tokio not native thread
    let stub_term = Arc::new(AtomicBool::new(false));

    let metrics = Metrics::new(&config);
    metrics.start(stub_term.clone());

    let (sub_results_sender, mut sub_results) = mpsc::channel::<ZmqMessage>(1_000_000);

    tokio::spawn(async move {
        loop {
            match sub_results.recv().await {
                Some(m) => {
                    socket_pub.send(m);
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
                        metrics.message_egress(); // TODO this would be better to have in receiving queue
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

// fn setup_logging(&self) -> Result<(), LifecycleError> {
// 	SimpleLogger::new().init()?;
//
// 	log::set_max_level(LevelFilter::Info);
//
// 	let level = match &*self.config.log_level {
// 		"DEBUG" => LevelFilter::Debug,
// 		"INFO" => LevelFilter::Info,
// 		"WARN" => LevelFilter::Warn,
// 		"ERROR" => LevelFilter::Error,
// 		_ => {
// 			log::warn!(
//                     "Invalid log level {}, using level INFO",
//                     self.config.log_level
//                 );
// 			LevelFilter::Info
// 		}
// 	};
//
// 	log::info!("Log level set to {}", level.as_str());
// 	log::set_max_level(level);
//
// 	Ok(())
// }

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
