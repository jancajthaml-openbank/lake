use config::Configuration;
use metrics::Metrics;
use std::sync::Arc;

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

    /// # Errors
    ///
    /// Propagates `zmq:Error`
    pub fn run(&self) -> Result<(), zmq::Error> {
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
            self.metrics.message_ingress();
            publisher.send(data, 0)?;
            self.metrics.message_egress();
        }
    }
}
