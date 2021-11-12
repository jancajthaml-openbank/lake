use std::fmt;
use std::process::exit;
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::mpsc::{channel, Sender};
use std::sync::Arc;
use std::thread;
use std::time::{Duration, SystemTime};

use log::info;
use statsd::Client;
use systemstat::{saturating_sub_bytes, DateTime, Platform, System, Utc};

use crate::config::Configuration;
use crate::metrics::MetricCmdType::{DUMP, EGRESS, INGRESS, TERM};

#[derive(Clone, Eq, PartialEq)]
pub enum MetricCmdType {
    DUMP,
    INGRESS,
    EGRESS,
    TERM,
}

/// statsd metrics subroutine
pub struct Metrics {
    pub sender: Sender<MetricCmdType>,
}

impl Drop for Metrics {
    fn drop(&mut self) {
        let _ = self.sender.send(TERM);
        // drop(&self.receiving_handle);
        // let _ = self.ticker_handler.abort();
        // drop(&self.ticker_handler);
        drop(&self.sender);
    }
}

impl Metrics {
    /// creates new metrics fascade
    #[must_use]
    pub fn new(config: &Configuration) -> Metrics {
        let (metrics_sender, metrics_receiver) = channel::<MetricCmdType>();
        info!("Setting up metrics");

        let s1 = metrics_sender.clone();
        let s2 = metrics_sender.clone();

        let messaging_handle = thread::spawn(move || {
            let duration = Duration::from_secs(1);
            loop {
                thread::sleep(duration);
                let _ = s1.send(DUMP);
            }
        });

        let endpoint: String = config.statsd_endpoint.clone();
        let client = match Client::new(&endpoint, "openbank.lake") {
            Ok(client) => client,
            Err(_) => {
                eprintln!("unable to initialise statsd client");
                exit(1); // TODO really exit?
            }
        };

        let messaging_handle = thread::spawn(move || {
            let mut ingress: u32 = 0;
            let mut egress: u32 = 0;

            let system = System::new();
            loop {
                match metrics_receiver.recv() {
                    Ok(cmd) if cmd == DUMP => {
                        send_metrics(&client, &system, &ingress, &egress);
                        ingress = 0;
                        egress = 0;
                    }
                    Ok(cmd) if cmd == INGRESS => {
                        ingress += 1;
                    }
                    Ok(cmd) if cmd == EGRESS => {
                        egress += 1;
                    }
                    Ok(cmd) if cmd == TERM => {
                        send_metrics(&client, &system, &ingress, &egress);
                        log::info!("TERMINATING metrics loop");
                        break;
                    }
                    Ok(_) => {
                        log::info!("OK_")
                    }
                    Err(_) => {
                        log::warn!("Err receiving");
                        break;
                    }
                }
            }
        });

        Metrics { sender: s2 }
    }
}

// send metrics to statsd client
#[allow(clippy::cast_precision_loss)]
fn send_metrics(client: &Client, system: &System, ingress: &u32, egress: &u32) {
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
