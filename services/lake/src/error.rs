use std::ffi;
use std::fmt;
use std::{mem, str};
use zmq_sys::errno;

#[allow(clippy::upper_case_acronyms)]
#[derive(Clone, Copy, Eq, PartialEq)]
pub enum Error {
    EACCES,
    EADDRINUSE,
    EAGAIN,
    EBUSY,
    ECONNREFUSED,
    EFAULT,
    EINTR,
    EHOSTUNREACH,
    EINPROGRESS,
    EINVAL,
    EMFILE,
    EMSGSIZE,
    ENAMETOOLONG,
    ENODEV,
    ENOENT,
    ENOMEM,
    ENOTCONN,
    ENOTSOCK,
    EPROTO,
    EPROTONOSUPPORT,
    ENOTSUP,
    ENOBUFS,
    ENETDOWN,
    EADDRNOTAVAIL,
    // native zmq error codes
    EFSM,
    ENOCOMPATPROTO,
    ETERM,
    EMTHREAD,
}

impl Error {
    pub fn to_raw(self) -> i32 {
        match self {
            Error::EACCES => errno::EACCES,
            Error::EADDRINUSE => errno::EADDRINUSE,
            Error::EAGAIN => errno::EAGAIN,
            Error::EBUSY => errno::EBUSY,
            Error::ECONNREFUSED => errno::ECONNREFUSED,
            Error::EFAULT => errno::EFAULT,
            Error::EINTR => errno::EINTR,
            Error::EHOSTUNREACH => errno::EHOSTUNREACH,
            Error::EINPROGRESS => errno::EINPROGRESS,
            Error::EINVAL => errno::EINVAL,
            Error::EMFILE => errno::EMFILE,
            Error::EMSGSIZE => errno::EMSGSIZE,
            Error::ENAMETOOLONG => errno::ENAMETOOLONG,
            Error::ENODEV => errno::ENODEV,
            Error::ENOENT => errno::ENOENT,
            Error::ENOMEM => errno::ENOMEM,
            Error::ENOTCONN => errno::ENOTCONN,
            Error::ENOTSOCK => errno::ENOTSOCK,
            Error::EPROTO => errno::EPROTO,
            Error::EPROTONOSUPPORT => errno::EPROTONOSUPPORT,
            Error::ENOTSUP => errno::ENOTSUP,
            Error::ENOBUFS => errno::ENOBUFS,
            Error::ENETDOWN => errno::ENETDOWN,
            Error::EADDRNOTAVAIL => errno::EADDRNOTAVAIL,

            Error::EFSM => errno::EFSM,
            Error::ENOCOMPATPROTO => errno::ENOCOMPATPROTO,
            Error::ETERM => errno::ETERM,
            Error::EMTHREAD => errno::EMTHREAD,
        }
    }

    pub fn from_raw(raw: i32) -> Error {
        match raw {
            errno::EACCES => Error::EACCES,
            errno::EADDRINUSE | errno::EADDRINUSE_ALT => Error::EADDRINUSE,
            errno::EAGAIN => Error::EAGAIN,
            errno::EBUSY => Error::EBUSY,
            errno::ECONNREFUSED | errno::ECONNREFUSED_ALT => Error::ECONNREFUSED,
            errno::EFAULT => Error::EFAULT,
            errno::EHOSTUNREACH | errno::EHOSTUNREACH_ALT => Error::EHOSTUNREACH,
            errno::EINPROGRESS | errno::EINPROGRESS_ALT => Error::EINPROGRESS,
            errno::EINVAL => Error::EINVAL,
            errno::EMFILE => Error::EMFILE,
            errno::EMSGSIZE | errno::EMSGSIZE_ALT => Error::EMSGSIZE,
            errno::ENAMETOOLONG => Error::ENAMETOOLONG,
            errno::ENODEV => Error::ENODEV,
            errno::ENOENT => Error::ENOENT,
            errno::ENOMEM => Error::ENOMEM,
            errno::ENOTCONN | errno::ENOTCONN_ALT => Error::ENOTCONN,
            errno::ENOTSOCK | errno::ENOTSOCK_ALT => Error::ENOTSOCK,
            errno::EPROTO => Error::EPROTO,
            errno::EPROTONOSUPPORT | errno::EPROTONOSUPPORT_ALT => Error::EPROTONOSUPPORT,
            errno::ENOTSUP | errno::ENOTSUP_ALT => Error::ENOTSUP,
            errno::ENOBUFS | errno::ENOBUFS_ALT => Error::ENOBUFS,
            errno::ENETDOWN | errno::ENETDOWN_ALT => Error::ENETDOWN,
            errno::EADDRNOTAVAIL | errno::EADDRNOTAVAIL_ALT => Error::EADDRNOTAVAIL,
            errno::EINTR => Error::EINTR,

            // 0MQ native error codes
            errno::EFSM => Error::EFSM,
            errno::ENOCOMPATPROTO => Error::ENOCOMPATPROTO,
            errno::ETERM => Error::ETERM,
            errno::EMTHREAD => Error::EMTHREAD,

            x => unsafe {
                let s = zmq_sys::zmq_strerror(x);
                panic!(
                    "unknown error [{}]: {}",
                    x,
                    str::from_utf8(ffi::CStr::from_ptr(s).to_bytes()).unwrap()
                )
            },
        }
    }

    /// Returns the error message provided by 0MQ.
    pub fn message(self) -> &'static str {
        unsafe {
            let s = zmq_sys::zmq_strerror(self.to_raw());
            let v: &'static [u8] = mem::transmute(ffi::CStr::from_ptr(s).to_bytes());
            str::from_utf8(v).unwrap()
        }
    }
}

impl From<Error> for String {
    fn from(error: Error) -> String {
        error.to_string()
    }
}

impl std::error::Error for Error {
    fn description(&self) -> &str {
        self.message()
    }
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.message())
    }
}

impl fmt::Debug for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.message())
    }
}
