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
    pub fn new(config: &Configuration, metrics: Arc<Metrics>) -> Relay {
        Relay {
            pull_port: config.pull_port,
            pub_port: config.pub_port,
            metrics,
            ctx: zmq::Context::new(),
        }
    }

    pub fn run(&self) -> Result<(), ()> {
        let puller = self.ctx.socket(zmq::PULL).unwrap();

        puller.set_immediate(true).unwrap();
        puller.set_conflate(false).unwrap();
        puller.set_linger(0).unwrap();
        puller.set_sndhwm(0).unwrap();

        let publisher = self.ctx.socket(zmq::PUB).unwrap();

        publisher.set_immediate(true).unwrap();
        publisher.set_conflate(false).unwrap();
        publisher.set_linger(0).unwrap();
        publisher.set_sndhwm(0).unwrap();

        puller
            .bind(&format!("tcp://127.0.0.1:{}", self.pull_port))
            .unwrap();
        publisher
            .bind(&format!("tcp://127.0.0.1:{}", self.pub_port))
            .unwrap();

        loop {
            let data = puller.recv_bytes(0).unwrap();
            self.metrics.message_ingress();
            publisher.send(data, 0).unwrap();
            self.metrics.message_egress();
        }
    }
}
