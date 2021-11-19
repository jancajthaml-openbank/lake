use colored::Colorize;
use log::{Level, LevelFilter, Log, Metadata, Record, SetLoggerError};
use std::time::SystemTime;

pub struct Logger {}

impl Logger {
    pub fn new() -> Logger {
        Logger {}
    }

    pub fn init(self, level: &str) -> Result<(), SetLoggerError> {
        log::set_max_level(LevelFilter::Info);
        log::set_boxed_logger(Box::new(self))?;

        let max_level = match level {
            "DEBUG" => LevelFilter::Debug,
            "INFO" => LevelFilter::Info,
            "WARN" => LevelFilter::Warn,
            "ERROR" => LevelFilter::Error,
            _ => {
                log::warn!("Invalid log level {}, using level INFO", level);
                LevelFilter::Info
            }
        };
        log::info!("Log level set to {}", max_level);
        log::set_max_level(max_level);

        Ok(())
    }
}

impl Log for Logger {
    fn enabled(&self, metadata: &Metadata) -> bool {
        metadata.level().to_level_filter() <= log::max_level()
    }

    fn log(&self, record: &Record) {
        if self.enabled(record.metadata()) {
            let level_string = match record.level() {
                Level::Error => "ERR".red().to_string(),
                Level::Warn => "WRN".red().to_string(),
                Level::Info => "INF".green().to_string(),
                Level::Debug => "DBG".yellow().to_string(),
                Level::Trace => "".to_owned(),
            };

            let target = if record.target().is_empty() {
                record.module_path().unwrap_or_default()
            } else {
                record.target()
            };

            let now = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap();

            let ts = now.as_secs();

            let mut year = 1970;

            let dayclock = ts % 86400;
            let mut dayno = ts / 86400;

            let sec = (dayclock % 60) as i32;
            let min = ((dayclock % 3600) / 60) as i32;
            let hour = (dayclock / 3600) as i32;

            loop {
                let yearsize = if leapyear(year) { 366 } else { 365 };
                if dayno >= yearsize {
                    dayno -= yearsize;
                    year += 1;
                } else {
                    break;
                }
            }
            year = year as i32;

            let mut monno = 0;
            while dayno >= MONTHS[if leapyear(year) { 1 } else { 0 }][monno] {
                dayno -= MONTHS[if leapyear(year) { 1 } else { 0 }][monno];
                monno += 1;
            }
            let month = monno as i32 + 1;
            let day = dayno as i32 + 1;

            println!(
                "{:04}-{:02}-{:02}T{:02}:{:02}:{:02}Z {} [{}] {}",
                year,
                month,
                day,
                hour,
                min,
                sec,
                level_string,
                target,
                record.args()
            );
        }
    }

    fn flush(&self) {}
}

fn leapyear(year: i32) -> bool {
    year % 4 == 0 && (year % 100 != 0 || year % 400 == 0)
}

static MONTHS: [[u64; 12]; 2] = [
    [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31],
    [31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31],
];
