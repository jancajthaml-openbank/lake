extern crate zmq;

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
            pull_port: *&config.pull_port,
            pub_port: *&config.pub_port,
            metrics: metrics,
            ctx: zmq::Context::new(),
        }
    }

    pub fn run(&self) {
        let puller = self.ctx.socket(zmq::PULL).unwrap();
        let publisher = self.ctx.socket(zmq::PUB).unwrap();

        puller.bind(&format!("tcp://127.0.0.1:{}", self.pull_port)).unwrap();
        publisher.bind(&format!("tcp://127.0.0.1:{}", self.pub_port)).unwrap();

        loop {
            let msg = puller.recv_msg(0).unwrap();
            self.metrics.message_ingress();
            publisher.send(msg, 0).unwrap();
            self.metrics.message_egress();
        }
    }
}
