use lazy_static::lazy_static;

use boot::Program;

fn main() {
    lazy_static! {
        static ref PROGRAM: Program = Program::new();
    }
    // FIXME have return Result<>
    PROGRAM.setup();
    // FIXME have return Result<>
    PROGRAM.start();
    // FIXME have return Result<>
    PROGRAM.stop();
}
