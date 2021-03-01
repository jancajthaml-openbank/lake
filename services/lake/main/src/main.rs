use boot::Program;

fn main() {
	let program = Program::new();
	program.setup();
	program.start();
	program.stop();
}