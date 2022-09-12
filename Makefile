DATA_DIR?=./data
OUTPUT_DIR?=./OUTPUT_DIR
CARGO_DEV_OPTIONS=--manifest-path=rust/Cargo.toml

run-rust-server:
	cargo run $(CARGO_DEV_OPTIONS) --release

run-go-server:
	cd go && go run main.go

clean: clean-rust
	rm -rf $(OUTPUT_DIR)

check: check-rust build-go

test: test-rust test-go

fmt: fmt-rust fmt-go

lint: lint-rust lint-go

clean-rust:
	cargo clean $(CARGO_DEV_OPTIONS)

check-rust:
	cargo check $(CARGO_DEV_OPTIONS)

test-rust:
	cargo test $(CARGO_DEV_OPTIONS)

fmt-rust:
	cargo fmt $(CARGO_DEV_OPTIONS)

lint-rust:
	cargo clippy $(CARGO_DEV_OPTIONS)

build-go:
	cd go && go build

test-go:
	cd go && go test

fmt-go:
	cd go && go fmt

lint-go:
	cd go && go vet
