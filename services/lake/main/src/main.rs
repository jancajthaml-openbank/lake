use lazy_static::lazy_static;

use boot::Program;

fn main() {
    lazy_static! {
        static ref PROGRAM: Program = Program::new();
    }
    if let Err(e) = PROGRAM.setup() {
        panic!(e);
    };
    if let Err(e) = PROGRAM.start() {
        panic!(e);
    };
    PROGRAM.stop();
}
