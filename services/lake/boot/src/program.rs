
use crate::health::{notify_service_ready, notify_service_stopping};

use config::Configuration;
use metrics::Metrics;
use relay::Relay;

use std::time::Duration;
use std::thread;
use log::LevelFilter;
use simple_logger::SimpleLogger;
use signal_hook::iterator::Signals;
use signal_hook::low_level;
use signal_hook::consts::{TERM_SIGNALS, SIGTERM, SIGQUIT};

pub struct Program {
	config: Configuration,
}

impl Program {

    pub fn new() -> Program {
        Program {
        	config: Configuration::load(),
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
    			log::warn!("Invalid log level {}, using level INFO", self.config.log_level);
    		    LevelFilter::Info
    		},
    	};

    	log::info!("Log level set to {}", level.as_str());
    	log::set_max_level(level);
    }

    pub fn setup(&self) {
    	self.setup_logging();
        log::info!("Program Setup");
    }

    pub fn start(&self) {
        log::info!("Program Starting");
        notify_service_ready();

        let mut pool = Vec::with_capacity(2);

        let metrics = Metrics::new(&self.config);
		let relay = Relay::new(&self.config, &metrics);

		pool.push(metrics.run());
		pool.push(relay.run());
        
		// https://github.com/bastion-rs/bastion

        //pool
        //let pool = thread::spawn(|| {

        	//for _ in 0..5 {
        	//	log::info!("program working...");
        		//thread::sleep(Duration::from_millis(1000));
        	//}
        //low_level::raise(SIGTERM).unwrap();
        //});

    	let mut sigs = Signals::new(TERM_SIGNALS).unwrap();
        for _ in sigs.forever() {
        	//_pool
        	pool.
            //break
        };

        pool.join();
        notify_service_stopping();
    }

    pub fn stop(&self) {
        log::info!("Program Stopping");
        low_level::raise(SIGQUIT).unwrap();
    }
}
