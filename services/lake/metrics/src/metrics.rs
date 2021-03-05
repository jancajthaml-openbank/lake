use config::Configuration;
use statsd::Client;
use std::fmt;
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::{Duration, SystemTime};
use systemstat::{saturating_sub_bytes, Platform, System};

pub struct Metrics {
    statsd_endpoint: String,
    ingress: Arc<AtomicU32>,
    egress: Arc<AtomicU32>,
}

impl Metrics {
    #[must_use]
    pub fn new(config: &Configuration) -> Metrics {
        Metrics {
            statsd_endpoint: config.statsd_endpoint.clone(),
            ingress: Arc::new(AtomicU32::new(0)),
            egress: Arc::new(AtomicU32::new(0)),
        }
    }

    pub fn message_egress(&self) {
        self.egress.fetch_add(1, Ordering::SeqCst);
    }

    pub fn message_ingress(&self) {
        self.ingress.fetch_add(1, Ordering::SeqCst);
    }

    #[must_use]
    pub fn start(&self, term_sig: Arc<AtomicBool>) -> std::thread::JoinHandle<()> {
        let statsd_endpoint = self.statsd_endpoint.clone();
        let ingress = self.ingress.clone();
        let egress = self.egress.clone();
        thread::spawn({
            move || {
                match Client::new(&statsd_endpoint, "openbank.lake") {
                    Ok(client) => {
                        let system = System::new();
                        let one_sec = Duration::from_secs(1);
                        let short_sleep = Duration::from_millis(100);
                        let mut last_time = SystemTime::now();
                        log::debug!("entering loop");
                        let mut alive = true;
                        while alive {
                            let now = SystemTime::now();
                            if let Ok(dur) = now.duration_since(last_time) {
                                if dur >= one_sec {
                                    last_time = now;
                                    send_metrics(&client, &system, &ingress, &egress);
                                    alive &= !term_sig.load(Ordering::Relaxed);
                                }
                            }
                            thread::sleep(short_sleep);
                        }
                        log::debug!("exiting loop");
                    }
                    Err(e) => {
                        log::warn!("unable to initialize statsd client with {:?}", e);
                        term_sig.store(true, Ordering::Relaxed);
                    }
                };
            }
        })
    }

    /// # Errors
    ///
    /// Yields `StopError` when failed to stop gracefully
    #[allow(clippy::unused_self)]
    pub fn stop(&self) -> Result<(), StopError> {
        log::debug!("requested stop");
        Ok(())
    }
}

#[allow(clippy::cast_precision_loss)]
fn send_metrics(
    client: &Client,
    system: &System,
    ingress: &Arc<AtomicU32>,
    egress: &Arc<AtomicU32>,
) {
    let mut pipe = client.pipeline();

    let v_ingress = ingress.load(Ordering::SeqCst);
    let v_egress = egress.load(Ordering::SeqCst);

    if let Ok(mem) = system.memory() {
        pipe.gauge(
            "memory.bytes",
            saturating_sub_bytes(mem.total, mem.free).as_u64() as f64,
        )
    }

    pipe.count("message.ingress", f64::from(v_ingress));
    pipe.count("message.egress", f64::from(v_egress));

    pipe.send(client);

    ingress.fetch_sub(v_ingress, Ordering::SeqCst);
    egress.fetch_sub(v_egress, Ordering::SeqCst);
}

pub struct StopError;

impl fmt::Display for StopError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "unable to stop metrics")
    }
}
