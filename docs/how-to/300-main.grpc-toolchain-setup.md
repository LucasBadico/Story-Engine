# gRPC Toolchain Setup

> Step-by-step guide to install the Protocol Buffers compiler and Go gRPC plugins on WSL2, then compile `.proto` files for the main service.

---

## Prerequisites

- WSL2 with Ubuntu 24.04
- Go 1.22+ installed and in PATH (`go version` works)
- `$GOPATH/bin` in PATH (check with `echo $PATH | grep go/bin`)
- `unzip` installed (`sudo apt update && sudo apt install unzip -y`)

---

## 1. Install protoc (Protocol Buffers Compiler)

protoc compiles `.proto` files into language-specific code. The apt version is usually outdated, so install from GitHub releases.

```bash
# Check latest at https://github.com/protocolbuffers/protobuf/releases
# As of March 2026, v29.x is current
PROTOC_VERSION=29.3
PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Download
wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}

# Install to /usr/local (protoc binary + well-known .proto includes)
sudo unzip -o ${PROTOC_ZIP} -d /usr/local bin/protoc 'include/*'

# Clean up
rm ${PROTOC_ZIP}
```

Verify:

```bash
protoc --version
# Expected: libprotoc 29.3 (or newer)
```

---

## 2. Install Go Protobuf Plugins

Two plugins are needed — one for message types (`protoc-gen-go`) and one for gRPC service stubs (`protoc-gen-go-grpc`).

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Verify both are in PATH:

```bash
which protoc-gen-go
# Expected: /home/<user>/go/bin/protoc-gen-go

which protoc-gen-go-grpc
# Expected: /home/<user>/go/bin/protoc-gen-go-grpc
```

If `which` returns nothing, add `$GOPATH/bin` to your PATH in `~/.bashrc`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Then `source ~/.bashrc`.

---

## 3. Compile Proto Files

From the main service directory:

```bash
cd main-service
```

### Compile all protos (existing Makefile target)

```bash
make proto-gen
```

This runs:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/**/*.proto
```

### Compile a single proto (useful when iterating)

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/character_state/character_state.proto
```

### What gets generated

For each `proto/<entity>/<entity>.proto`, two files are created:

| File | Contains |
|------|----------|
| `<entity>.pb.go` | Message types (structs), marshal/unmarshal |
| `<entity>_grpc.pb.go` | gRPC client + server interfaces, `Register<Service>Server()` |

### Clean generated files

```bash
make proto-clean
# or: find proto -name "*.pb.go" -delete
```

---

## 4. Compile the New Proto Files (Task 1)

The 4 new protos that need compilation:

```bash
cd main-service

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/character_state/character_state.proto \
       proto/relationship_state/relationship_state.proto \
       proto/timeline_event/timeline_event.proto \
       proto/plot/plot.proto
```

After this, each directory will have the generated `*.pb.go` and `*_grpc.pb.go` files.

### Next steps after compilation

1. Create gRPC handler files in `internal/transport/grpc/` (one per service)
2. Implement each RPC by calling the existing use cases
3. Register services in `cmd/api-grpc/main.go`

Pattern to follow: look at any existing handler in `internal/transport/grpc/` — they all follow the same structure (embed `Unimplemented*Server`, call use case, map domain ↔ proto).

---

## 5. Verification Checklist

```bash
# 1. protoc installed
protoc --version
# libprotoc 29.3+

# 2. Go plugins installed
which protoc-gen-go
# /home/<user>/go/bin/protoc-gen-go

which protoc-gen-go-grpc
# /home/<user>/go/bin/protoc-gen-go-grpc

# 3. Compilation works
cd main-service && make proto-gen
# No errors, proto/**/*_grpc.pb.go files exist

# 4. Build still passes
go build ./...
# No errors

# 5. New protos compiled
ls proto/character_state/*.pb.go
# character_state.pb.go  character_state_grpc.pb.go
```

All green? You're ready to write gRPC handlers.

---

## Quick Reference

| What | Command |
|------|---------|
| Install protoc | `wget + sudo unzip` (see Step 1) |
| Install Go plugins | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` + `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest` |
| Compile all protos | `cd main-service && make proto-gen` |
| Compile one proto | `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/<entity>/<entity>.proto` |
| Clean generated code | `make proto-clean` |
| Run gRPC server | `make run-grpc` |

---

## Troubleshooting

### `protoc: command not found`
Binary not in PATH. Check it exists at `/usr/local/bin/protoc`. If you installed elsewhere, symlink or add to PATH.

### `protoc-gen-go: program not found or is not executable`
The Go plugin isn't in PATH. Run `which protoc-gen-go` — if empty, ensure `$GOPATH/bin` is in your PATH (see Step 2).

### `google/protobuf/timestamp.proto: File not found`
The well-known proto includes weren't installed. Re-run the `sudo unzip` step from Step 1 — the `include/*` part extracts these files to `/usr/local/include/google/protobuf/`.

### `proto/**/*.proto: No such file or directory`
Glob expansion failed. You need to be in the `main-service/` directory, or the shell doesn't support `**` glob. Use bash 4+ or run `shopt -s globstar` first.

### Generated code has import errors
Run `go mod tidy` in `main-service/`. The proto files require `google.golang.org/protobuf` and `google.golang.org/grpc` — both should already be in `go.mod`.

---

## WSL2 Known Issues

### DNS breaks randomly (`no such host`, `dial tcp: lookup ... on 100.100.100.100:53`)

WSL2 sometimes loses DNS resolution — `go install`, `apt`, `wget`, and anything that needs network will fail with host lookup errors.

**Fix** (run every time it happens):

```bash
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
```

You can chain it before any network command to be safe:

```bash
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

This points DNS at Google's public resolver. It works immediately, no restart needed.

**Do NOT** edit `/etc/wsl.conf` with `generateResolvConf = false` — this has broken SSH access in the past. The one-liner above is the safe fix.

### `unzip: command not found`

Fresh Ubuntu on WSL2 may not have `unzip` pre-installed:

```bash
sudo apt update && sudo apt install unzip -y
```

### PostgreSQL not running after WSL restart

WSL2 doesn't auto-start services. After each `wsl --shutdown` or reboot:

```bash
sudo service postgresql start
```
