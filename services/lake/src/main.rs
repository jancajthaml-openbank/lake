mod metrics;
mod config;

use std::process::exit;
use std::sync::Arc;
use zmq::Message;
use std::{env, io};
use std::os::unix::net::UnixDatagram;
use tokio::io::Error;
use crate::config::Configuration;
use crate::metrics::Metrics;
use log::{info, warn, debug, error, LevelFilter};

#[tokio::main(flavor = "multi_thread")]
// #[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(),Error> {

	println!("STARTING ABC");
	ready();
	let metrics = Metrics::new(&Configuration::load());
	let ctx = zmq::Context::new();

	println!("STARTING B");
	let mut socket_pull = ctx.socket(zmq::PULL).unwrap();
	socket_pull.set_conflate(false);
	socket_pull.set_immediate(true);
	socket_pull.set_linger(0);
	socket_pull.set_rcvhwm(0);
	socket_pull.connect("tcp://127.0.0.1:5562").unwrap();

	println!("STARTING C");
	let mut socket_pub = ctx.socket(zmq::PUB).unwrap();
	socket_pub.connect("tcp://127.0.0.1:5561").unwrap();
	socket_pub.set_conflate(false);
	socket_pub.set_immediate(true);
	socket_pub.set_linger(0);
	socket_pub.set_rcvhwm(0);

	println!("STARTING D");

	// setup_logging()
	debug!("Starting");
	info!("Starting");
	warn!("Starting");
	error!("Starting");


	let mut msg: Message = Message::new();
	loop {
		match socket_pull.recv(&mut msg, 0) {
			Ok(_) => {
				metrics.message_ingress();
				// tokio::spawn(async move {
				// 	socket_pub.send(msg, 0);
				metrics.message_egress();
				// });
			}
			Err(e) => {
				eprintln!("{}", e);
				stopping();
				exit(0);
			}
		}
	}
	println!(">>> END");
	Result::Err(Error::from_raw_os_error(1))
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
