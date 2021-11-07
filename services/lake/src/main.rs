mod config;
mod metrics;

use crate::config::Configuration;
use crate::metrics::Metrics;
use std::os::unix::net::UnixDatagram;
use std::process::exit;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;
use std::{env, io};
use tokio::io::Error;
use zeromq::*;

//#[tokio::main(flavor = "multi_thread")]
#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), Error> {
    // FIXME coordinated shutdown https://tokio.rs/tokio/topics/shutdown

    //println!("STARTING C");
    //println!("STARTING ABC");

    let config = Configuration::load();
    let metrics = Metrics::new(&config);

    // STUB for setup logging and bbtests
    println!("Log level set to {}", config.log_level);

    //let mut pull = zeromq::RepSocket::new();
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

    ready();

    // FIXME better with tokio not native thread
    let stub_term = Arc::new(AtomicBool::new(false));
    metrics.start(stub_term);

    // current thoughtput ~100k msg/sec

    loop {
        tokio::select! {
            message = socket_pull.recv() => {
                metrics.message_ingress();
                socket_pub
                    .send(message)
                    .await
                    .expect("Failed to bind SEND message");
                metrics.message_egress();
            },
            // FIXME add term signal handler for coordinated shutdown
        };
    }
    //println!(">>> END");
    //Result::Err(Error::from_raw_os_error(1))
    //Ok(())
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
