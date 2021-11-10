use std::fmt;
use std::process::exit;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::thread;
use std::time::{Duration, SystemTime};

use statsd::Client;
use systemstat::{DateTime, Platform, saturating_sub_bytes, System, Utc};
use tokio::{task, time};
use tokio::sync::{broadcast, mpsc};
use tokio::sync::mpsc::{channel, Sender};
use tokio::task::JoinHandle;

use crate::config::Configuration;
use crate::metrics::MetricCmdType::{DUMP, EGRESS, INGRESS};
use log::info;
use tokio::time::sleep;

#[derive(Clone, Eq, PartialEq)]
pub enum MetricCmdType {
	DUMP,
	INGRESS,
	EGRESS,
}

/// statsd metrics subroutine
pub struct Metrics {
	// dump_handle: JoinHandle<String>,
	receiving_handle: JoinHandle<String>,
	pub sender: Sender<MetricCmdType>,
}

impl Metrics {
	/// creates new metrics fascade
	#[must_use]
	pub fn new(config: &Configuration) -> Metrics {
		let (metrics_sender, mut metrics_receiver) = mpsc::channel::<MetricCmdType>(10_000);
		info!("Setting up metrics");

		let s = metrics_sender.clone();
		let s2 = metrics_sender.clone();
		let handle1 = tokio::spawn(async move {		// TODO convert to thread
			loop {
				sleep(Duration::from_secs(10)).await;
				let _ = s.send(MetricCmdType::DUMP).await;
			}
		});

		let endpoint: String = config.statsd_endpoint.clone();
		let handle2 = tokio::spawn(async move {		// TODO convert to thread
			let mut ingress: u32 = 0;
			let mut egress: u32 = 0;
			let system = System::new();
			loop {
				match metrics_receiver.recv().await {
					Some(cmd) if cmd == DUMP => {
						let now = SystemTime::now();
						let now: DateTime<Utc> = now.into();
						info!("Metrics dump {} -> {}/{}", now.to_rfc3339(), ingress, egress);
						// match Client::new(&endpoint, "openbank.lake") {
						// 	Ok(client) => {
						// 		send_metrics2(&client, &system, &ingress, &egress);
						// 		ingress = 0;  // TODO reset counter in sending success or always?
						// 		egress = 0;
						// 	}
						// 	Err(e) => eprintln!("{}", e)
						// }
					}
					Some(cmd) if cmd == INGRESS => {
						ingress += 1;
					}
					Some(cmd) if cmd == EGRESS => {
						egress += 1;
					}
					Some(_) => {
						log::info!("OK_")
					}
					None => {
						log::warn!("Err receiving",);
						exit(1);
					}
				}
			}
		});

		Metrics {
			// dump_handle: handle1,
			receiving_handle: handle2,
			sender: s2,
		}
	}

	/// increments egress counter
	pub async fn message_egress(&self) {
		let _ = self.sender.send(EGRESS).await;
	}

	/// increments ingress counter
	pub async fn message_ingress(&self) {
		let _ = self.sender.send(INGRESS).await;
	}

	/// # Errors
	///
	/// yields `StopError` when failed to stop gracefully
	#[allow(clippy::unused_self)]
	pub fn stop(&self) -> Result<(), StopError> {
		log::debug!("requested stop");
		Ok(())
	}
}

// send metrics to statsd client
#[allow(clippy::cast_precision_loss)]
fn send_metrics2(
	client: &Client,
	system: &System,
	ingress: &u32,
	egress: &u32,
) {
	let mut pipe = client.pipeline();

	if let Ok(mem) = system.memory() {
		pipe.gauge(
			"memory.bytes",
			saturating_sub_bytes(mem.total, mem.free).as_u64() as f64,
		)
	}

	pipe.count("message.ingress", *ingress as _);
	pipe.count("message.egress", *egress as _);

	pipe.send(client);
}

pub struct StopError;

impl fmt::Display for StopError {
	fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
		write!(f, "unable to stop metrics")
	}
}
