#[macro_use]
extern crate lazy_static;

use boot::Program;

fn main() {
	lazy_static! {
        static ref PROGRAM: Program = Program::new();
    }
	PROGRAM.setup();
	PROGRAM.start();
	PROGRAM.stop();
}