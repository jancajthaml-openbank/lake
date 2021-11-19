use libc::{c_void, size_t};
use std::ffi;
use std::mem;

use crate::error;

pub struct Context {
    pub underlying: *mut c_void,
}
unsafe impl Send for Context {}
unsafe impl Sync for Context {}

impl Context {
    pub fn new() -> Context {
        Context {
            underlying: unsafe { zmq_sys::zmq_ctx_new() },
        }
    }

    pub fn set_io_threads(&self, value: usize) -> Result<(), error::Error> {
        if unsafe {
            zmq_sys::zmq_ctx_set(
                self.underlying,
                zmq_sys::ZMQ_IO_THREADS as i32,
                value as i32,
            )
        } == -1
        {
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }
}

impl Drop for Context {
    fn drop(&mut self) {
        log::debug!("Stopping ZMQ context");
        while unsafe { zmq_sys::zmq_ctx_term(self.underlying) } == -1 {
            if error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }) != error::Error::EINTR {
                break;
            };
        }
        log::debug!("Stopped ZMQ context");
    }
}

pub struct Socket {
    pub sock: *mut c_void,
}

impl Socket {
    pub fn new(ctx: *mut c_void, socket_type: u32) -> Result<Socket, error::Error> {
        let sock = unsafe { zmq_sys::zmq_socket(ctx, socket_type as i32) };
        if sock.is_null() {
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(Socket { sock })
    }

    pub fn bind(&self, endpoint: &str) -> Result<(), error::Error> {
        let cstr = ffi::CString::new(endpoint.as_bytes()).unwrap();
        if unsafe { zmq_sys::zmq_bind(self.sock, cstr.as_ptr()) } == -1 {
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }

    pub fn connect(&self, endpoint: &str) -> Result<(), error::Error> {
        let cstr = ffi::CString::new(endpoint.as_bytes()).unwrap();
        if unsafe { zmq_sys::zmq_connect(self.sock, cstr.as_ptr()) } == -1 {
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }

    pub fn set_option(&self, opt: u32, val: i32) -> Result<(), error::Error> {
        if unsafe {
            zmq_sys::zmq_setsockopt(
                self.sock,
                opt as i32,
                (&val as *const i32).cast::<c_void>(),
                mem::size_of::<i32>() as size_t,
            )
        } == -1
        {
            return Err(error::Error::from_raw(unsafe { zmq_sys::zmq_errno() }));
        }
        Ok(())
    }
}

impl Drop for Socket {
    fn drop(&mut self) {
        log::debug!("Stopping ZMQ socket");
        if unsafe { zmq_sys::zmq_close(self.sock) } == -1 {
            log::error!("socket leaked");
        }
        log::debug!("Stopped ZMQ socket");
    }
}
