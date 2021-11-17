use signal_hook::consts::SIGQUIT;
use signal_hook::low_level;
use statsd::Client;
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;

use crate::config::Configuration;

/// statsd metrics subroutine
pub struct Metrics {
    /// ingress counter
    ingress: Arc<AtomicUsize>,
    /// egress counter
    egress: Arc<AtomicUsize>,
    // FIXME needs child thread handle
    child_thread: Option<thread::JoinHandle<()>>,
}

impl Drop for Metrics {
    fn drop(&mut self) {
        log::info!("Metrics stopping");
        log::debug!("Metrics waiting for child thread to terminate");
        let _ = self.child_thread.take().unwrap().join();
        log::info!("Metrics stopped");
    }
}

impl Metrics {
    /// creates new metrics fascade
    #[must_use]
    pub fn new(config: &Configuration, prog_running: Arc<AtomicBool>) -> Arc<Metrics> {
        let endpoint: String = config.statsd_endpoint.clone();

        let arc_ingress = Arc::new(AtomicUsize::new(0));
        let arc_ingress_clone = arc_ingress.clone();

        let arc_egress = Arc::new(AtomicUsize::new(0));
        let arc_egress_clone = arc_egress.clone();

        log::info!("Metrics starting");

        let child_thread = thread::spawn(move || {
            let statsd_client = match Client::new(&endpoint, "openbank.lake") {
                Ok(client) => Some(client),
                Err(_) => {
                    log::error!("unable to initialise statsd client");
                    None
                }
            };

            let duration = Duration::from_secs(1);

            match statsd_client {
                Some(statsd_client) => {
                    log::info!("Metrics started");
                    while prog_running.load(Ordering::Relaxed) {
                        thread::sleep(duration);

                        let mut pipe = statsd_client.pipeline();

                        pipe.gauge("memory.bytes", mem_bytes());
                        pipe.count(
                            "message.ingress",
                            arc_ingress_clone.swap(0, Ordering::Relaxed) as _,
                        );
                        pipe.count(
                            "message.egress",
                            arc_egress_clone.swap(0, Ordering::Relaxed) as _,
                        );

                        pipe.send(&statsd_client);
                    }
                }
                _ => {}
            }

            let _ = low_level::raise(SIGQUIT);
        });

        Arc::new(Metrics {
            ingress: arc_ingress.clone(),
            egress: arc_egress.clone(),
            child_thread: Some(child_thread),
        })
    }

    /// increments egress counter
    pub fn message_egress(&self) {
        self.egress.fetch_add(1, Ordering::Relaxed);
    }

    /// increments ingress counter
    pub fn message_ingress(&self) {
        self.ingress.fetch_add(1, Ordering::Relaxed);
    }
}

#[cfg(target_os = "linux")]
fn mem_bytes() -> f64 {
    if let Ok(me) = procfs::process::Process::myself() {
        if let Ok(page_size) = procfs::page_size() {
            return (me.stat.rss * page_size) as f64;
        };
    };
    0 as f64
}

#[cfg(target_os = "macos")]
fn mem_bytes() -> f64 {
    0 as f64
}
