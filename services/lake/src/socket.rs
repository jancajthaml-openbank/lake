use libc::{c_int, c_void, size_t};
use std::ffi;
use std::mem;
use zmq_sys;

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
}

pub struct Socket {
    pub sock: *mut c_void,
}

impl Socket {
    pub fn new(ctx: *mut c_void, t: c_int) -> Result<Socket, error::Error> {
        let sock = unsafe { zmq_sys::zmq_socket(ctx, t) };
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

    pub fn set_option(&self, opt: c_int, val: i32) -> Result<(), error::Error> {
        if unsafe {
            zmq_sys::zmq_setsockopt(
                self.sock,
                opt,
                (&val as *const i32) as *const c_void,
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
        if unsafe { zmq_sys::zmq_close(self.sock) } == -1 {
            panic!("socket leaked!");
        }
    }
}
