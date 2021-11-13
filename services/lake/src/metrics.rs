use std::sync::mpsc::{channel, Sender};
use std::thread;
use std::time::Duration;

use log::info;
use statsd::Client;
//use systemstat::{saturating_sub_bytes, Platform, System};

use signal_hook::consts::SIGQUIT;
use signal_hook::low_level;

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
        drop(&self.sender);
    }
}

impl Metrics {
    /// creates new metrics fascade
    #[must_use]
    pub fn new(config: &Configuration) -> Result<Metrics, String> {
        let (metrics_sender, metrics_receiver) = channel::<MetricCmdType>();
        info!("Setting up metrics");

        let s1 = metrics_sender.clone();
        let s2 = metrics_sender.clone();

        thread::spawn(move || {
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
                let _ = low_level::raise(SIGQUIT);
                return Err("unable to initialise statsd client".to_owned());
            }
        };

        thread::spawn(move || {
            let mut ingress: u32 = 0;
            let mut egress: u32 = 0;

            loop {
                match metrics_receiver.recv() {
                    Ok(cmd) if cmd == DUMP => {
                        send_metrics(&client, &ingress, &egress);
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
                        send_metrics(&client, &ingress, &egress);
                        log::info!("TERMINATING metrics loop");
                        break;
                    }
                    Ok(_) => {}
                    Err(_) => {
                        break;
                    }
                }
            }

            let _ = low_level::raise(SIGQUIT);
        });

        Ok(Metrics { sender: s2 })
    }
}

// send metrics to statsd client
#[allow(clippy::cast_precision_loss)]
fn send_metrics(client: &Client, ingress: &u32, egress: &u32) {
    let mut pipe = client.pipeline();
    if let Ok(me) = procfs::process::Process::myself() {
        pipe.gauge("memory.bytes", me.stat.vsize as f64)
    }

    pipe.count("message.ingress", *ingress as _);
    pipe.count("message.egress", *egress as _);

    pipe.send(client);
}
