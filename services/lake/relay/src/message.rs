//use std::error::Error;
use std::fmt;
use std::ops::{Deref, DerefMut};
use std::slice;
use zmq;

/// Holds a ZMQ message.
pub struct Message {
    msg: zmq_sys::zmq_msg_t,
}

impl Drop for Message {
    fn drop(&mut self) {
        //log::debug!("dropping message");
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
            panic!(zmq::Error::from_raw(zmq_sys::zmq_errno()))
        }
        Message { msg }
    }

    pub fn new() -> Message {
        //log::debug!("allocating message");
        unsafe { Self::alloc(|msg| zmq_sys::zmq_msg_init(msg)) }
    }

    /*
    pub fn isempty(&self) -> bool {
        let x = unsafe { zmq_sys::zmq_msg_size(&self.msg) };
        x == 0
    }
    */
}

/*
impl Deref for Message {
    type Target = [u8];

    fn deref(&self) -> &[u8] {
        // This is safe because we're constraining the slice to the lifetime of
        // this message.
        unsafe {
            let ptr = &self.msg as *const _ as *mut _;
            let data = zmq_sys::zmq_msg_data(ptr);
            let len = zmq_sys::zmq_msg_size(ptr) as usize;
            slice::from_raw_parts(data as *mut u8, len)
        }
    }
}*/

//impl Eq for Message {}

/*
impl DerefMut for Message {
    fn deref_mut(&mut self) -> &mut [u8] {
        // This is safe because we're constraining the slice to the lifetime of
        // this message.
        unsafe {
            let data = zmq_sys::zmq_msg_data(&mut self.msg);
            let len = zmq_sys::zmq_msg_size(&self.msg) as usize;
            slice::from_raw_parts_mut(data as *mut u8, len)
        }
    }
}*/

/// Get the low-level C pointer.
pub fn msg_ptr(msg: &mut Message) -> *mut zmq_sys::zmq_msg_t {
    &mut msg.msg
}
