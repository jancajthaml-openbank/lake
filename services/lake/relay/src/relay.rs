use crate::message::{msg_ptr, Message};

use config::Configuration;
use metrics::Metrics;
use std::error::Error;
use std::fmt;
//use std::ops::Deref;
use std::ffi;
use std::mem;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;

//use zmq_sys::libc::c_int;
use libc::{c_int, c_void, size_t};

macro_rules! zmq_try {
    ($($tt:tt)*) => {{
        let rc = $($tt)*;
        if rc == -1 {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        rc
    }}
}

struct Context {
    underlying: *mut c_void,
}
unsafe impl Send for Context {}
unsafe impl Sync for Context {}

impl Context {
    fn new() -> Context {
        Context {
            underlying: unsafe { zmq_sys::zmq_ctx_new() },
        }
    }
}

struct Socket {
    sock: *mut c_void,
    // The `context` field is never accessed, but implicitly does
    // reference counting via the `Drop` trait.
    //#[allow(dead_code)]
    //context: Option<Context>,
    //owned: bool,
}

//unsafe impl Send for Socket {}

impl Socket {
    fn new(ctx: *mut c_void, t: c_int) -> Result<Socket, zmq::Error> {
        let sock = unsafe { zmq_sys::zmq_socket(ctx, t) };
        if sock.is_null() {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(Socket { sock })
    }

    fn bind(&self, endpoint: &str) -> Result<(), zmq::Error> {
        let cstr = ffi::CString::new(endpoint.as_bytes()).unwrap();
        if unsafe { zmq_sys::zmq_bind(self.sock, cstr.as_ptr()) } == -1 {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }

    fn set_option(&self, opt: c_int, val: i32) -> Result<(), zmq::Error> {
        if unsafe {
            zmq_sys::zmq_setsockopt(
                self.sock,
                opt,
                (&val as *const i32) as *const c_void,
                mem::size_of::<i32>() as size_t,
            )
        } == -1
        {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }
}
impl Drop for Socket {
    fn drop(&mut self) {
        if unsafe { zmq_sys::zmq_close(self.sock) } == -1 {
            panic!("!");
        }
    }
}
/// message relay subroutine
pub struct Relay {
    /// port to bind ZMQ PULL
    pull_port: i32,
    /// port to bind ZMQ PUB
    pub_port: i32,
    /// instance of `Metrics`
    metrics: Arc<Metrics>,
    /// ZMQ context
    ctx: Arc<Context>,
}

impl Relay {
    #[must_use]
    pub fn new(config: &Configuration, metrics: Arc<Metrics>) -> Relay {
        Relay {
            pull_port: config.pull_port,
            pub_port: config.pub_port,
            metrics,
            ctx: Arc::new(Context::new()),
        }
    }

    /// starts thread relaying messages
    #[must_use]
    pub fn start(&self, term_sig: Arc<AtomicBool>) -> std::thread::JoinHandle<()> {
        let ctx = self.ctx.clone();
        let metrics = self.metrics.clone();
        let endpoint_pull = format!("tcp://127.0.0.1:{}", self.pull_port);
        let endpoint_pub = format!("tcp://127.0.0.1:{}", self.pub_port);
        thread::spawn({
            move || {
                log::debug!("entering loop");
                while !term_sig.load(Ordering::Relaxed) {
                    if let Err(e) = pull_to_pub(&ctx, &metrics, &endpoint_pull, &endpoint_pub) {
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
    /// yields `StopError` when failed to stop gracefully
    pub fn stop(&self) -> Result<(), StopError> {
        log::debug!("requested stop");
        log::debug!("terminating ZMQ context");
        if unsafe { zmq_sys::zmq_ctx_term(self.ctx.underlying) } == -1 {
            log::debug!("failed terminating ZMQ context");
            return Err(StopError::from(zmq::Error::from_raw(unsafe {
                zmq_sys::zmq_errno()
            })));
        };
        log::debug!("done terminating ZMQ context");
        Ok(())
    }
}

/// relays all ZMQ PULL messages to ZMQ PUB
/// on message received from ZMQ PULL report message ingress to metrics
/// on message received from ZMQ PUB report message egress to metrics
///
/// # Errors
///
/// propagates `zmq:Error` on empty message circuit breaks with ETERM
fn pull_to_pub(
    ctx: &Arc<Context>,
    metrics: &Arc<Metrics>,
    endpoint_pull: &str,
    endpoint_pub: &str,
) -> Result<(), zmq::Error> {
    let puller = Socket::new(ctx.underlying, zmq_sys::ZMQ_PULL as c_int)?;
    puller.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0)?;
    puller.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1)?;
    puller.set_option(zmq_sys::ZMQ_LINGER as c_int, 0)?;
    puller.set_option(zmq_sys::ZMQ_RCVHWM as c_int, 0)?;
    puller.bind(endpoint_pull)?;
    // INFO does not unbinds on drop

    let publisher = Socket::new(ctx.underlying, zmq_sys::ZMQ_PUB as c_int)?;
    publisher.set_option(zmq_sys::ZMQ_CONFLATE as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_IMMEDIATE as c_int, 1)?;
    publisher.set_option(zmq_sys::ZMQ_LINGER as c_int, 0)?;
    publisher.set_option(zmq_sys::ZMQ_SNDHWM as c_int, 0)?;
    publisher.bind(endpoint_pub)?;
    // INFO does not unbinds on drop

    // FIXME inline
    loop {
        let mut msg = Message::new();
        let ptr = msg_ptr(&mut msg);
        if unsafe { zmq_sys::zmq_msg_recv(ptr, puller.sock, 0 as c_int) } == -1 {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        //metrics.message_ingress(); // FIXME this seems to be slowing the code from 23s to 53s (1M -> 500k / sec)
        if unsafe {
            let data = zmq_sys::zmq_msg_data(ptr);
            let len = zmq_sys::zmq_msg_size(ptr) as usize;
            zmq_sys::zmq_send(publisher.sock, data, len, 0 as c_int)
        } == -1
        {
            return Err(zmq::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        };
        //metrics.message_egress(); // FIXME this seems to be slowing the code from 23s to 53s (1M -> 500k / sec)
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
