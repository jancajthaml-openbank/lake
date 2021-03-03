use config::Configuration;
use metrics::Metrics;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Barrier};
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

    pub fn start(
        &'static self,
        term_sig: Arc<AtomicBool>,
        barrier: Arc<Barrier>,
    ) -> std::thread::JoinHandle<()> {
        log::info!("start relay");
        thread::spawn({
            let term = term_sig.clone();
            move || {
                while !term.load(Ordering::Relaxed) {
                    if let Err(e) = self.work() {
                        log::warn!("relay recovering from crash {:?}", e);
                    }
                }
                barrier.wait();
                log::debug!("relay exiting loop");
            }
        })
    }

    pub fn stop(&self) {
        log::debug!("requested stop relay");

        let kill_message = zmq::Message::new();
        let killer = self.ctx.socket(zmq::PUSH).unwrap();

        killer
            .connect(&format!("tcp://127.0.0.1:{}", self.pull_port))
            .unwrap();

        killer.send(kill_message, 0).unwrap();
    }

    /// # Errors
    ///
    /// Propagates `zmq:Error`
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
            if data.len() == 0 {
                return Err(zmq::Error::ETERM);
            }
            self.metrics.message_ingress();
            publisher.send(data, 0)?;
            self.metrics.message_egress();
        }
    }
}
