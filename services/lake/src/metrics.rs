use std::fmt;
use std::process::exit;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::thread;
use std::time::{Duration, SystemTime};

use statsd::Client;
use systemstat::{Platform, saturating_sub_bytes, System};
use tokio::{task, time};
use tokio::sync::{broadcast, mpsc};
use tokio::sync::broadcast::{channel, Sender};

use crate::config::Configuration;
use crate::metrics::MetricCmdType::{DUMP, EGRESS, INGRESS};

#[derive(Clone, Eq, PartialEq)]
pub enum MetricCmdType {
	DUMP,
	INGRESS,
	EGRESS,
}

/// statsd metrics subroutine
pub struct Metrics {
	pub sender: Sender<MetricCmdType>,
}

impl Metrics {
	/// creates new metrics fascade
	#[must_use]
	pub fn new(config: &Configuration) -> Metrics {
		let (metrics_sender, _metrics_receiver) = broadcast::channel::<MetricCmdType>(10_000);

		let s = metrics_sender.clone();
		let s2 = metrics_sender.clone();
		tokio::spawn(async move {		// TODO convert to thread
			let mut interval = time::interval(Duration::from_secs(1));
			loop {
				interval.tick().await;
				let _ = s.send(MetricCmdType::DUMP);
			}
		});

		let endpoint: String = config.statsd_endpoint.clone();
		tokio::spawn(async move {		// TODO convert to thread
			let mut ingress: u32 = 0;
			let mut egress: u32 = 0;
			let system = System::new();
			loop {
				match metrics_sender.clone().subscribe().recv().await {
					Ok(cmd) if cmd == DUMP => {
						// println!("Duping metrics");
						match Client::new(&endpoint, "openbank.lake") {
							Ok(client) => {
								send_metrics2(&client, &system, &ingress, &egress);
								ingress = 0;  // TODO reset counter in sending success or always?
								egress = 0;
							}
							Err(e) => eprintln!("{}", e)
						}
					}
					Ok(cmd) if cmd == INGRESS => {
						ingress += 1;
					}
					Ok(cmd) if cmd == EGRESS => {
						egress += 1;
					}
					Ok(_) => {
						log::info!("OK_")
					}
					Err(e) => {
						log::warn!("Err {}", e);
						exit(1);
					}
				}
			}
		});

		Metrics {
			sender: s2,
		}
	}

	/// increments egress counter
	pub fn message_egress(&self) {
		let _ = self.sender.send(EGRESS);
	}

	/// increments ingress counter
	pub fn message_ingress(&self) {
		let _ = self.sender.send(INGRESS);
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
