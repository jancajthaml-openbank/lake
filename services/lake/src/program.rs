use crate::config::Configuration;
use log::LevelFilter;
use signal_hook::consts::{SIGQUIT, TERM_SIGNALS};
use signal_hook::iterator::Signals;
use signal_hook::low_level;
use simple_logger::SimpleLogger;
use std::error::Error;
use std::fmt;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
//use crate::metrics;
//use crate::metrics::Metrics;

pub struct Program {
    /// configuration
    config: Configuration,
    // statsd metrics
    // metrics: Arc<Metrics>,
    // message relay
    // relay: Arc<Relay>,
}

impl Program {
    /// initializes program
    #[must_use]
    pub fn new() -> Program {
        let config = Configuration::load();
        // let metrics = Arc::new(Metrics::new(&config));
        // let relay = Arc::new(Relay::new(&config, Arc::clone(&metrics)));

        Program {
            config,
            // metrics,
            // relay,
        }
    }

    /// setups logging from configuration
    /// on invalid log level falls back to log level info
    fn setup_logging(&self) -> Result<(), LifecycleError> {
        SimpleLogger::new().init()?;

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

        Ok(())
    }

    /// initializes program
    ///
    /// # Errors
    ///
    /// yields `LifecycleError` when failed to setup
    pub fn setup(&self) -> Result<(), LifecycleError> {
        self.setup_logging()?;
        log::info!("Program Setup");
        Ok(())
    }

    /// starts all program subroutines, interuptable by terminal
    /// signals
    ///
    /// # Errors
    ///
    /// yields `LifecycleError` when failed to start
    pub fn start(&self) -> Result<(), LifecycleError> {
        log::info!("Program Starting");

        let term_now = Arc::new(AtomicBool::new(false));

        // let metrics_handle = self.metrics.start(term_now.clone());
        // let relay_handle = self.relay.start(term_now.clone());

        log::info!("Program Started");

        let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
        let _ = sigs.wait();
        log::info!("signal received, going down");
        term_now.store(true, Ordering::Relaxed);

        // self.metrics.stop()?;
        // self.relay.stop()?;

        // metrics_handle.join()?;
        // relay_handle.join()?;

        log::info!("Program Stopping");

        Ok(())
    }

    /// sends `SIGQUIT` to program
    #[allow(clippy::unused_self)]
    pub fn stop(&self) {
        low_level::raise(SIGQUIT).unwrap();
    }
}

impl Default for Program {
    fn default() -> Self {
        Program::new()
    }
}

#[derive(Debug)]
pub struct LifecycleError {
    details: String,
}

impl Error for LifecycleError {
    fn description(&self) -> &str {
        &self.details
    }
}

impl LifecycleError {
    fn new(msg: &str) -> LifecycleError {
        LifecycleError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for LifecycleError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

//impl From<metrics::StopError> for LifecycleError {
//  fn from(err: metrics::StopError) -> Self {
//    LifecycleError::new(&err.to_string())
//}
//}

// impl From<relay::StopError> for LifecycleError {
//     fn from(err: relay::StopError) -> Self {
//         LifecycleError::new(&err.to_string())
//     }
// }

impl From<log::SetLoggerError> for LifecycleError {
    fn from(err: log::SetLoggerError) -> Self {
        LifecycleError::new(&err.to_string())
    }
}

impl From<std::boxed::Box<dyn std::any::Any + std::marker::Send>> for LifecycleError {
    fn from(err: std::boxed::Box<dyn std::any::Any + std::marker::Send>) -> Self {
        err.downcast_ref::<String>().map_or_else(
            || LifecycleError::new("runtime panic"),
            |s| LifecycleError::new(s),
        )
    }
}
