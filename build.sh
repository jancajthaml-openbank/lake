
cargo build \
	--release \
	--manifest-path=services/lake/Cargo.toml \
	--target-dir=/private/tmp/rust-lake

/private/tmp/rust-lake/release/main