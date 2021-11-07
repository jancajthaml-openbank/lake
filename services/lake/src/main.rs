mod config;
mod metrics;

use crate::config::Configuration;
use crate::metrics::Metrics;
use std::net::SocketAddr;
use std::os::unix::net::UnixDatagram;
use std::process::exit;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;
use std::{env, io};
use tokio::io::Error;
use tokio::net::TcpListener;
use tokio::net::TcpStream;
use tokio::signal::unix::{signal, SignalKind};
use tokio::sync::mpsc;

// FIXME ideally remove in future
use zeromq::*;

//#[tokio::main(flavor = "multi_thread")]
#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), Error> {
    let config = Configuration::load();
    let metrics = Metrics::new(&config);

    // STUB for setup logging and bbtests
    println!("Log level set to {}", config.log_level);

    // >>> INFO low level straight forward implementation START
    // FIXME missing unbind on termination / drop ?
    let pull_listener = TcpListener::bind(("127.0.0.1", 2250 as u16))
        .await
        .expect("Failed to bind RAW TCP PULL socket");

    // Coordinated shutdown section
    // FIXME is this broacast for all coroutines (greenlets) or is this simple queue?
    let (shutdown_order, shutdown_listen) = mpsc::unbounded_channel::<bool>();

    let mut term_sig = signal(SignalKind::terminate()).unwrap();

    tokio::spawn(async move {
        loop {
            tokio::select! {
                _ = term_sig.recv() => {
                    println!("TERM signal received");
                    shutdown_order.send(true);
                    break
                }
            }
        }
    });

    // Accept from PULL
    tokio::spawn(async move {
        let mut stop = shutdown_listen;
        ready();
        println!("Loop Started");
        loop {
            tokio::select! {
                incoming = pull_listener.accept() => {
                    let maybe_accepted: Result<_, _> = incoming.and_then(|(raw_socket, remote_addr)|{
                        raw_socket.set_nodelay(true).map(|_| { (raw_socket, remote_addr) })
                    })
                    .map(|(raw_socket, remote_addr)| {
                        // FIXME maybe don't need to frame socket connection, need to decode it with codec first
                        // this is PULL socket, when no monitor is in ZMQ topology it should only SEND messages
                        // no need for multiplexing maybe
                        let (read, write) = tokio::io::split(raw_socket);

                        // INFO ZMQ connection is keep-alive

                        // INFO there is FSM when ZMQ connection is accepted
                        // Greeting -> FrameHeader
                        // FrameHeader -> FrameLength
                        // FrameLength -> Frame
                        // Frame -> FrameHeader

                        let framed = (Box::new(read), Box::new(write));
                        //let endpoint = SocketAddr::new(remote_addr.ip(), remote_addr.port());
                        (framed, remote_addr)
                    });
                    //.map_err(|err| err.into());

                    // FIXME missing ZMQ codec right now, not sure if TCP is keep-alive or for each message rn.
                    //println!("accepted TCP socket on PULL");

                    // FIXME eventually we should see "YXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXZ"
                    // messages accepted here via ZMQ PUSH from perf testing
                },
                _ = stop.recv() => {
                    println!("Loop Stopped");
                    stopping();
                    std::process::exit(0);
                    break
                }
            }
        }
        //Ok(())
    });
    // <<< INFO low level straight forward implementation END

    /////////////////////////////////
    // Reference implementation below
    /////////////////////////////////

    let mut socket_pull = zeromq::PullSocket::new();
    socket_pull
        .bind(&format!("tcp://127.0.0.1:{}", config.pull_port)) // FIXME temporarily -1 for above tcp accept
        .await
        .expect("Failed to bind PULL socket");

    let mut socket_pub = zeromq::PubSocket::new();
    socket_pub
        .bind(&format!("tcp://127.0.0.1:{}", config.pub_port))
        .await
        .expect("Failed to bind PUB socket");

    // FIXME better with tokio not native thread
    let stub_term = Arc::new(AtomicBool::new(false));
    metrics.start(stub_term);

    // current thoughtput ~100k msg/sec

    loop {
        tokio::select! {
            message = socket_pull.recv() => {
                metrics.message_ingress();
                socket_pub
                    .send(message.unwrap())
                    .await
                    .expect("Failed to bind SEND message");
                metrics.message_egress();
            },
            // FIXME add term signal handler for coordinated shutdown
        };
    }

    //println!(">>> END");
    //Result::Err(Error::from_raw_os_error(1))
    //Ok(())

    /*
    code snippets what does this loop actualy do in pseudo code using zmq.rs library

    loop {
        loop {
            match pull_buffer.next() {
                Some(message) => {
                    for subsriber in pub_subsribers {
                        subscriber.clone().queue.clone().try_send(Message(message.clone))
                    }
                }
                None => todo!(),
            };
        }
    }

    */
}

// fn setup_logging(&self) -> Result<(), LifecycleError> {
// 	SimpleLogger::new().init()?;
//
// 	log::set_max_level(LevelFilter::Info);
//
// 	let level = match &*self.config.log_level {
// 		"DEBUG" => LevelFilter::Debug,
// 		"INFO" => LevelFilter::Info,
// 		"WARN" => LevelFilter::Warn,
// 		"ERROR" => LevelFilter::Error,
// 		_ => {
// 			log::warn!(
//                     "Invalid log level {}, using level INFO",
//                     self.config.log_level
//                 );
// 			LevelFilter::Info
// 		}
// 	};
//
// 	log::info!("Log level set to {}", level.as_str());
// 	log::set_max_level(level);
//
// 	Ok(())
// }

/// tries to notify host os that service is ready
fn ready() {
    if let Err(e) = notify("READY=1") {
        println!("unable to notify host os about READY with {}", e);
    }
}

/// tries to notify host os that service is stopping
fn stopping() {
    if let Err(e) = notify("STOPPING=1") {
        println!("unable to notify host os about STOPPING with {}", e)
    }
}

/// sends msg to `NOTIFY_SOCKET` via udp
fn notify(msg: &str) -> io::Result<()> {
    let socket_path = match env::var_os("NOTIFY_SOCKET") {
        Some(path) => path,
        None => return Ok(()),
    };
    let sock = UnixDatagram::unbound()?;
    let len = sock.send_to(msg.as_bytes(), socket_path)?;
    if len == msg.len() {
        Ok(())
    } else {
        Err(io::Error::new(io::ErrorKind::WriteZero, "incomplete write"))
    }
}
