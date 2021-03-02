use config::Configuration;
use statsd::Client;
use std::sync::atomic::AtomicU32;
use std::sync::atomic::Ordering;
use systemstat::{saturating_sub_bytes, Platform, System};

pub struct Metrics {
    client: Client,
    ingress: AtomicU32,
    egress: AtomicU32,
    system: System,
}

impl Metrics {
    #[must_use]
    pub fn new(config: &Configuration) -> Metrics {
        Metrics {
            system: System::new(),
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
        let mut pipe = self.client.pipeline();

        let ingress = self.ingress.load(Ordering::SeqCst);
        let egress = self.egress.load(Ordering::SeqCst);

        match self.system.memory() {
            Ok(mem) => pipe.gauge(
                "memory.bytes",
                saturating_sub_bytes(mem.total, mem.free).as_u64() as f64,
            ),
            Err(_) => (),
        }

        pipe.count("message.ingress", f64::from(ingress));
        pipe.count("message.egress", f64::from(egress));

        pipe.send(&self.client);

        self.ingress.fetch_sub(ingress, Ordering::SeqCst);
        self.egress.fetch_sub(egress, Ordering::SeqCst);
    }
}
