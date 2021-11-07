mod metrics;
mod config;

use std::process::exit;
use std::sync::Arc;
use zmq::Message;
use std::io;
use tokio::io::Error;
use crate::config::Configuration;
use crate::metrics::Metrics;
use log::{info, warn, debug, error, LevelFilter};

#[tokio::main(flavor = "multi_thread")]
// #[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(),Error> {

	let metrics = Metrics::new(&Configuration::load());
	let ctx = zmq::Context::new();

	let mut socket_pull = ctx.socket(zmq::PULL).unwrap();
	socket_pull.set_conflate(false);
	socket_pull.set_immediate(true);
	socket_pull.set_linger(0);
	socket_pull.set_rcvhwm(0);
	socket_pull.connect("tcp://127.0.0.1:5562").unwrap();

	let mut socket_pub = ctx.socket(zmq::PUB).unwrap();
	socket_pub.connect("tcp://127.0.0.1:5561").unwrap();
	socket_pub.set_conflate(false);
	socket_pub.set_immediate(true);
	socket_pub.set_linger(0);
	socket_pub.set_rcvhwm(0);

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
