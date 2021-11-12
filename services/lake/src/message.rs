use std::fmt;
use std::ops::{Deref, DerefMut};
use std::slice;
//use zmq;
use zmq_sys;

use crate::error;

/// Holds a ZMQ message.
pub struct Message {
    msg: zmq_sys::zmq_msg_t,
}

impl Drop for Message {
    fn drop(&mut self) {
        unsafe {
            let rc = zmq_sys::zmq_msg_close(&mut self.msg);
            assert_eq!(rc, 0);
        }
    }
}

impl Message {
    unsafe fn alloc<F>(f: F) -> Message
    where
        F: FnOnce(&mut zmq_sys::zmq_msg_t) -> i32,
    {
        let mut msg = zmq_sys::zmq_msg_t::default();
        let rc = f(&mut msg);
        if rc == -1 {
            panic!(error::Error::from_raw(zmq_sys::zmq_errno()))
        }
        Message { msg }
    }

    pub fn new() -> Message {
        unsafe { Self::alloc(|msg| zmq_sys::zmq_msg_init(msg)) }
    }

    /*
    pub fn isempty(&self) -> bool {
        let x = unsafe { zmq_sys::zmq_msg_size(&self.msg) };
        x == 0
    }
    */
}

/// Get the low-level C pointer.
pub fn msg_ptr(msg: &mut Message) -> *mut zmq_sys::zmq_msg_t {
    &mut msg.msg
}
