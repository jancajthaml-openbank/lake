use signal_hook::consts::SIGQUIT;
use signal_hook::low_level;
use statsd::Client;
use std::sync::atomic::AtomicUsize;
use std::sync::atomic::Ordering::Relaxed;
use std::sync::Arc;
use std::thread;
use std::time::Duration;

use crate::config::Configuration;

/// statsd metrics subroutine
pub struct Metrics {
    /// statsd client
    client: Client,
    /// ingress counter
    ingress: AtomicUsize,
    /// egress counter
    egress: AtomicUsize,
    // FIXME needs child thread handle
}

impl Drop for Metrics {
    fn drop(&mut self) {
        println!("Metrics stopping");
        log::info!("Metrics stopping");
        // FIXME need to join child thread here and wait
        self.send_metrics();
    }
}

impl Metrics {
    /// creates new metrics fascade
    #[must_use]
    pub fn new(config: &Configuration) -> Result<Arc<Metrics>, String> {
        let endpoint: String = config.statsd_endpoint.clone();
        let client = match Client::new(&endpoint, "openbank.lake") {
            Ok(client) => client,
            Err(_) => {
                let _ = low_level::raise(SIGQUIT);
                return Err("unable to initialise statsd client".to_owned());
            }
        };

        let instance = Metrics {
            client: client,
            ingress: AtomicUsize::new(0),
            egress: AtomicUsize::new(0),
        };

        let arc_instance = Arc::new(instance);
        let arc_instance_clone = Arc::clone(&arc_instance);

        log::info!("Metrics starting");

        // https://newbedev.com/what-happens-to-a-detached-thread-when-main-exits
        // https://github.com/rust-lang/rust/blob/8c069ceba81a0fffc1ce95aaf7e8339e11bf2796/src/libstd/sys/unix/thread.rs#L195

        // https://users.rust-lang.org/t/pthread-cancel-undefined-behavior/38477/4

        thread::spawn(move || {
            let duration = Duration::from_secs(1); // FIXME shorter sleeps
            loop {
                // FIXME check if program is dead

                // https://docs.rs/shuteye/0.3.3/shuteye/fn.sleep.html
                thread::sleep(duration);
                arc_instance_clone.send_metrics();
            }
        });

        Ok(arc_instance)
    }

    /// increments egress counter
    pub fn message_egress(&self) {
        self.egress.fetch_add(1, Relaxed);
    }

    /// increments ingress counter
    pub fn message_ingress(&self) {
        self.ingress.fetch_add(1, Relaxed);
    }

    fn send_metrics(&self) {
        let mut pipe = self.client.pipeline();

        pipe.gauge("memory.bytes", mem_bytes());
        pipe.count("message.ingress", self.ingress.swap(0, Relaxed) as _);
        pipe.count("message.egress", self.egress.swap(0, Relaxed) as _);

        pipe.send(&self.client);
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
