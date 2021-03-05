use config::Configuration;
use statsd::Client;
use std::fmt;
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::{Arc, Barrier};
use std::thread;
use std::time::Duration;
use systemstat::{saturating_sub_bytes, Platform, System};

pub struct Metrics {
    client: Client, // FIXME option
    system: System,
    ingress: AtomicU32,
    egress: AtomicU32,
    handle: Option<std::thread::JoinHandle<()>>,
}

impl Metrics {
    #[must_use]
    pub fn new(config: &Configuration) -> Metrics {
        Metrics {
            client: Client::new(&config.statsd_endpoint, "openbank.lake").unwrap(), // FIXME None right now
            system: System::new(),
            ingress: AtomicU32::new(0),
            egress: AtomicU32::new(0),
            handle: None,
        }
    }

    pub fn message_egress(&self) {
        self.egress.fetch_add(1, Ordering::SeqCst);
    }

    pub fn message_ingress(&self) {
        self.ingress.fetch_add(1, Ordering::SeqCst);
    }

    pub fn start(&'static mut self, term_sig: Arc<AtomicBool>, barrier: Arc<Barrier>) {
        log::info!("requested start");
        self.handle = Some(self.work(term_sig, barrier));
    }

    fn work(
        &'static self,
        term_sig: Arc<AtomicBool>,
        barrier: Arc<Barrier>,
    ) -> std::thread::JoinHandle<()> {
        thread::spawn({
            move || {
                while !term_sig.load(Ordering::Relaxed) {
                    // FIXME kill this thread / timer with stop
                    thread::sleep(Duration::from_secs(1));
                    self.send();
                }
                barrier.wait();
                log::debug!("exiting loop");
            }
        })
    }

    /// # Errors
    ///
    /// Yields `StopError` when failed to stop gracefully
    #[allow(clippy::unused_self)]
    pub fn stop(&self) -> Result<(), StopError> {
        log::debug!("requested stop");
        // FIXME terminate timer
        // and then
        //self.send();
        Ok(())
    }

    #[allow(clippy::cast_precision_loss)]
    fn send(&self) {
        let mut pipe = self.client.pipeline();

        let ingress = self.ingress.load(Ordering::SeqCst);
        let egress = self.egress.load(Ordering::SeqCst);

        if let Ok(mem) = self.system.memory() {
            pipe.gauge(
                "memory.bytes",
                saturating_sub_bytes(mem.total, mem.free).as_u64() as f64,
            )
        }

        pipe.count("message.ingress", f64::from(ingress));
        pipe.count("message.egress", f64::from(egress));

        pipe.send(&self.client);

        self.ingress.fetch_sub(ingress, Ordering::SeqCst);
        self.egress.fetch_sub(egress, Ordering::SeqCst);
    }
}

impl Drop for Metrics {
    fn drop(&mut self) {
        if let Some(handle) = self.handle.take() {
            handle.join().expect("failed to join thread");
        }
    }
}

pub struct StopError;

impl fmt::Display for StopError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "unable to stop metrics")
    }
}
