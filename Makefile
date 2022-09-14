DATA_TMP_DIR?=./data/tmp
OUTPUT_DIR?=./output
HOST?=localhost
TARGET?=slow
CARGO_DEV_OPTIONS=--manifest-path=rust/Cargo.toml

bench: ensure-bench-tool
	$(OUTPUT_DIR)/bin/oha -c 100 -z 30s http://$(HOST):8000/$(TARGET)

run-rust-server: ensure-data
	cargo run $(CARGO_DEV_OPTIONS) --release

run-go-server: ensure-data
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

ensure-output:
	mkdir -p $(OUTPUT_DIR)

ensure-data:
	mkdir -p $(DATA_TMP_DIR)

ensure-bench-tool: ensure-output
	cargo install oha --root $(OUTPUT_DIR)
	chmod +x $(OUTPUT_DIR)/bin/oha