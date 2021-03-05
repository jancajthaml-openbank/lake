use config::Configuration;
use metrics::Metrics;
use std::error::Error;
use std::fmt;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc; //{Arc, Barrier};
use std::thread;

pub struct Relay {
    pull_port: i32,
    pub_port: i32,
    metrics: Arc<Metrics>,
    ctx: zmq::Context,
}

impl Relay {
    #[must_use]
    pub fn new(config: &Configuration, metrics: Arc<Metrics>) -> Relay {
        Relay {
            pull_port: config.pull_port,
            pub_port: config.pub_port,
            metrics,
            ctx: zmq::Context::new(),
        }
    }

    #[must_use]
    pub fn start(&'static self, term_sig: Arc<AtomicBool>) -> std::thread::JoinHandle<()> {
        thread::spawn({
            move || {
                log::debug!("entering loop");
                while !term_sig.load(Ordering::Relaxed) {
                    if let Err(e) = self.work() {
                        log::warn!("crash {:?}", e);
                    }
                }
                log::debug!("exiting loop");
            }
        })
        // FIXME recover panic and set term_sig to upstream
    }

    /// # Errors
    ///
    /// Yields `StopError` when failed to stop gracefully
    pub fn stop(&self) -> Result<(), StopError> {
        log::debug!("requested stop");
        let kill_message = zmq::Message::new();
        let kill_sock = self.ctx.socket(zmq::PUSH)?; //.unwrap();
        kill_sock
            .connect(&format!("tcp://127.0.0.1:{}", self.pull_port))
            .unwrap();

        kill_sock.send(kill_message, 0)?;
        Ok(())
    }

    /// # Errors
    ///
    /// Propagates `zmq:Error` on empty message circuit breaks with ETERM
    fn work(&self) -> Result<(), zmq::Error> {
        let puller = self.ctx.socket(zmq::PULL)?;

        puller.set_immediate(true)?;
        puller.set_conflate(false)?;
        puller.set_linger(0)?;
        puller.set_sndhwm(0)?;

        let publisher = self.ctx.socket(zmq::PUB)?;

        publisher.set_immediate(true)?;
        publisher.set_conflate(false)?;
        publisher.set_linger(0)?;
        publisher.set_sndhwm(0)?;

        puller.bind(&format!("tcp://127.0.0.1:{}", self.pull_port))?;
        publisher.bind(&format!("tcp://127.0.0.1:{}", self.pub_port))?;

        loop {
            let data = puller.recv_bytes(0)?;
            if data.is_empty() {
                return Err(zmq::Error::ETERM);
            }
            self.metrics.message_ingress();
            publisher.send(data, 0)?;
            self.metrics.message_egress();
        }
    }
}

#[derive(Debug)]
pub struct StopError {
    details: String,
}

impl Error for StopError {
    fn description(&self) -> &str {
        &self.details
    }
}

impl StopError {
    fn new(msg: &str) -> StopError {
        StopError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for StopError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl From<zmq::Error> for StopError {
    fn from(err: zmq::Error) -> Self {
        StopError::new(&err.to_string())
    }
}
