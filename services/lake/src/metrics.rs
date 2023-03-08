use statsd::Client;
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;

use crate::config::Configuration;

/// statsd metrics subroutine
pub struct Metrics {
    /// messages counter
    messages: Arc<AtomicUsize>,
    // child thread join handle
    child_thread: Option<thread::JoinHandle<()>>,
}

impl Drop for Metrics {
    fn drop(&mut self) {
        log::info!("Metrics stopping");
        log::debug!("Metrics waiting for child thread to terminate");
        let _res = self.child_thread.take().unwrap().join();
        log::info!("Metrics stopped");
    }
}

impl Metrics {
    /// creates new metrics fascade
    #[must_use]
    #[allow(clippy::cast_precision_loss)]
    pub fn new(config: &Configuration, prog_running: Arc<AtomicBool>) -> Arc<Metrics> {
        let endpoint: String = config.statsd_endpoint.clone();

        let arc_messages = Arc::new(AtomicUsize::new(0));
        let arc_messages_clone = arc_messages.clone();

        log::info!("Metrics starting");

        let child_thread = thread::spawn(move || {
            let statsd_client = if let Ok(client) = Client::new(&endpoint, "openbank.lake") {
                Some(client)
            } else {
                log::error!("unable to initialise statsd client");
                None
            };

            let duration = Duration::from_secs(1);

            if let Some(statsd_client) = statsd_client {
                log::info!("Metrics started");
                while prog_running.load(Ordering::Relaxed) {
                    thread::sleep(duration);

                    let mut pipe = statsd_client.pipeline();

                    pipe.gauge("memory.bytes", mem_bytes());
                    pipe.count(
                        "message.relayed",
                        arc_messages_clone.swap(0, Ordering::Relaxed) as f64,
                    );

                    pipe.send(&statsd_client);
                }
            }
            unsafe { libc::raise(libc::SIGTERM) };
        });

        Arc::new(Metrics {
            messages: arc_messages,
            child_thread: Some(child_thread),
        })
    }

    /// increments ingress and egress counter
    pub fn relayed(&self) {
        self.messages.fetch_add(1, Ordering::Relaxed);
    }
}

#[cfg(target_os = "linux")]
#[allow(clippy::cast_precision_loss)]
#[inline]
fn mem_bytes() -> f64 {
    if let Ok(me) = procfs::process::Process::myself() {
        if let Ok(stat) = me.stat() {
            return (stat.rss * procfs::page_size()) as f64;
        };
    };
    0_f64
}

#[cfg(target_os = "macos")]
#[inline]
fn mem_bytes() -> f64 {
    0_f64
}
