extern crate statsd;

use statsd::Client;
use config::Configuration;
use std::sync::atomic::AtomicU32;
use std::sync::atomic::Ordering;
use sysinfo::{System, SystemExt};


pub struct Metrics {
    client: Client,
    ingress: AtomicU32,
    egress: AtomicU32,
}

impl Metrics {

    pub fn new(config: &Configuration) -> Metrics {
        Metrics {
            client: Client::new(&config.statsd_endpoint, "openbank.lake").unwrap(),
            ingress: AtomicU32::new(0),
            egress: AtomicU32::new(0),
        }
    }

    pub fn message_egress(&self) {
        self.egress.fetch_add(1, Ordering::SeqCst);
    }

    pub fn message_ingress(&self) {
        self.ingress.fetch_add(1, Ordering::SeqCst);
    }

    pub fn send(&self) {
        let mut system = System::new();
        let mut pipe = self.client.pipeline();

        let ingress = self.ingress.load(Ordering::SeqCst);
        let egress = self.egress.load(Ordering::SeqCst);
        system.refresh_memory();
        let memory = system.get_used_memory() as f64;

        pipe.count("message.ingress", ingress as f64);
        pipe.count("message.egress", egress as f64);
        pipe.gauge("memory.bytes", memory * 1000.0f64);

        pipe.send(&self.client);

        self.ingress.fetch_sub(ingress, Ordering::SeqCst);
        self.egress.fetch_sub(egress, Ordering::SeqCst);
    }
}
