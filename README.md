# IO benchmark

IO/Network benchmark comparison between different languages in different cases.

## Prepare

Install toolchain:

- Go
- Rust (cargo)
- [nghttp2](https://github.com/nghttp2/nghttp2)

## Start server

Start the target server you would like to test.

### Go

```bash
> make run-go-server
```

### Rust

```bash
> make run-rust-server
```

## Run benchmark

```bash
> make bench
```

### Run on remote

```bash
> HOST=192.168.1.110 make bench
```

### Run for fast request

```bash
> TARGET=fast make bench
```
