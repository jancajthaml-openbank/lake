use crate::health::{notify_service_ready, notify_service_stopping};

use config::Configuration;
use log::LevelFilter;
use metrics::Metrics;
use relay::Relay;
use signal_hook::consts::{SIGQUIT, TERM_SIGNALS};
use signal_hook::iterator::Signals;
use signal_hook::low_level;
use simple_logger::SimpleLogger;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Barrier};

pub struct Program {
    config: Configuration,
    metrics: Arc<Metrics>,
    relay: Arc<Relay>,
    barrier: Arc<Barrier>,
}

impl Program {
    #[must_use]
    pub fn new() -> Program {
        let config = Configuration::load();
        let metrics = Arc::new(Metrics::new(&config));
        let relay = Arc::new(Relay::new(&config, Arc::clone(&metrics)));

        Program {
            config,
            metrics,
            relay,
            barrier: Arc::new(Barrier::new(3)),
        }
    }

    fn setup_logging(&self) {
        SimpleLogger::new().init().unwrap();

        log::set_max_level(LevelFilter::Info);

        let level = match &*self.config.log_level {
            "DEBUG" => LevelFilter::Debug,
            "INFO" => LevelFilter::Info,
            "WARN" => LevelFilter::Warn,
            "ERROR" => LevelFilter::Error,
            _ => {
                log::warn!(
                    "Invalid log level {}, using level INFO",
                    self.config.log_level
                );
                LevelFilter::Info
            }
        };

        log::info!("Log level set to {}", level.as_str());
        log::set_max_level(level);
    }

    pub fn setup(&'static self) {
        self.setup_logging();
        log::info!("Program Setup");
    }

    pub fn start(&'static self) {
        log::info!("Program Starting");
        notify_service_ready();

        let term_now = Arc::new(AtomicBool::new(false));

        self.metrics.start(term_now.clone(), self.barrier.clone());
        self.relay.start(term_now.clone(), self.barrier.clone());

        let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
        let _ = sigs.wait();
        log::info!("signal received, stopping");
        term_now.store(true, Ordering::Relaxed);

        log::info!("coordinated shutdown");

        self.metrics.stop();
        self.relay.stop();
    }

    #[allow(clippy::unused_self)]
    pub fn stop(&'static self) {
        log::info!("Program Stopping");
        low_level::raise(SIGQUIT).unwrap();
        self.barrier.wait();
        log::info!("Program Stopped");
        notify_service_stopping();
    }
}

impl Default for Program {
    fn default() -> Self {
        Program::new()
    }
}
