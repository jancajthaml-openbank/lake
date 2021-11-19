use std::mem;
use std::sync::atomic::Ordering;

use crate::config::Configuration;
use crate::metrics::Metrics;
use crate::program::Program;
use crate::relay::Relay;

mod config;
mod error;
mod logger;
mod message;
mod metrics;
mod program;
mod relay;
mod socket;

fn main() {
    let config = Configuration::load();

    let prog = Program::new(&config);
    let metrics = Metrics::new(&config, prog.running.clone());
    let relay = Relay::new(&config, prog.running.clone(), metrics.clone());

    wait_for_eintr();

    prog.running.clone().store(false, Ordering::Relaxed);

    drop(relay);
    drop(metrics);
    drop(prog);
}

fn wait_for_eintr() {
    unsafe {
        let blocking = handler as extern fn(libc::c_int) as *mut libc::c_void as libc::sighandler_t;
        libc::signal(libc::SIGTERM, blocking);
    }

    let mut sigset_memspace = mem::MaybeUninit::uninit();
    let mut sigset: libc::sigset_t = unsafe {
        libc::sigfillset(sigset_memspace.as_mut_ptr());
        sigset_memspace.assume_init()
    };

    unsafe {
        libc::sigaddset(&mut sigset, libc::SIGTERM);
    };

    let mut signum = mem::MaybeUninit::uninit();
    unsafe {
        libc::sigwait(&sigset, signum.as_mut_ptr());
        signum.assume_init();
    };
}

extern fn handler(_: libc::c_int) {}