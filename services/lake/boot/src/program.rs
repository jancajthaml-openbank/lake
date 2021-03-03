use crate::health::{notify_service_ready, notify_service_stopping};

use config::Configuration;
use metrics::Metrics;
use relay::Relay;
use std::thread;

use bastion::prelude::*;
use log::LevelFilter;
use signal_hook::consts::{SIGQUIT, TERM_SIGNALS};
use signal_hook::iterator::Signals;
use signal_hook::low_level;
use simple_logger::SimpleLogger;
use std::sync::Arc;
use std::time::Duration;

pub struct Program {
    config: Configuration,
    metrics: Arc<Metrics>,
    relay: Arc<Relay>,
}

impl Program {
    
    #[must_use] pub fn new() -> Program {
        let config = Configuration::load();
        let metrics = Arc::new(Metrics::new(&config));
        let relay = Arc::new(Relay::new(&config, Arc::clone(&metrics)));

        Program {
            config,
            metrics,
            relay,
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

        Bastion::init();
        Bastion::start();

        Bastion::supervisor(|sp| {
            let callbacks = Callbacks::new()
                .with_before_start(|| log::debug!("Supervisor started."))
                .with_after_stop(|| log::debug!("Supervisor stopped."));

            sp.with_callbacks(callbacks)
                .with_strategy(SupervisionStrategy::OneForOne)
                .children(|children| {
                    let metrics = &self.metrics;

                    let callbacks = Callbacks::new().with_after_stop(move || metrics.send());

                    children
                        .with_callbacks(callbacks)
                        .with_exec(move |_ctx| async move {
                            loop {
                                thread::sleep(Duration::from_secs(1));
                                metrics.send();
                            }
                        })
                })
                .children(|children| {
                    let relay = &self.relay;
                    children
                        .with_exec(move |_ctx| async move {
                            relay.run().map_err(|e| {
                                log::warn!("relay crashed {:?}", e);
                            })
                        })
                })
                .children(|children| {
                    children.with_exec(|ctx| async move {
                        let mut sigs = Signals::new(TERM_SIGNALS).unwrap();

                        let sig = sigs.wait();
                        log::info!("signal {:?} received, stopping", sig);
                        ctx.parent()
                            .stop()
                            .expect("Couldn't stop signal group");
                        Bastion::stop();
                        log::info!("signal exit exec");
                        Ok(())
                    })
                })
        })
        .expect("Couldn't create the supervisor.");

        Bastion::block_until_stopped();
        notify_service_stopping();
    }

    #[allow(clippy::unused_self)]
    pub fn stop(&self) {
        log::info!("Program Stopping");
        low_level::raise(SIGQUIT).unwrap();
    }
}

impl Default for Program {
    fn default() -> Self {
        Program::new()
    }
}
