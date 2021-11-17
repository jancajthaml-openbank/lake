use signal_hook::consts::TERM_SIGNALS;
use signal_hook::iterator::Signals;
use std::sync::atomic::Ordering;

use crate::config::Configuration;
use crate::metrics::Metrics;
use crate::program::Program;
use crate::relay::Relay;

mod config;
mod error;
mod message;
mod metrics;
mod program;
mod relay;
mod socket;

fn main() -> Result<(), String> {
    let config = Configuration::load();

    let prog = Program::new(&config);
    let metrics = Metrics::new(&config, prog.running.clone());
    let relay = Relay::new(&config, prog.running.clone(), metrics.clone());

    let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
    let _ = sigs.wait();

    prog.running.clone().store(false, Ordering::Relaxed);

    drop(relay);
    drop(metrics);
    drop(prog);

    Ok(())
}
