# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SpacetimeDB is a relational database with embedded application logic (modules). Clients connect directly to the database via WebSocket/HTTP — there is no separate app server. Modules are uploaded as WebAssembly binaries compiled from Rust, C#, C++, or TypeScript. Each reducer call is a full ACID transaction. Clients subscribe to SQL queries and receive real-time delta updates.

## Build & Test Commands

### Building

```bash
# Default members (CLI, standalone server, updater)
cargo build

# All workspace crates
cargo build --all
```

### Testing

```bash
# CI test suite (Rust + C# tests, excludes smoketests)
cargo ci test

# All Rust tests directly (exclude smoketests)
cargo test --all --exclude spacetimedb-smoketests --exclude spacetimedb-sdk -- --test-threads=2

# Single crate
cargo test -p spacetimedb-core

# SDK tests (requires feature flag)
cargo test -p spacetimedb-sdk --features allow_loopback_http_for_tests

# Durability tests (must be serial)
cargo test -p spacetimedb-durability --features fallocate -- --test-threads=1

# Smoketests (rebuilds CLI+standalone first, then runs)
cargo smoketest

# Specific smoketest
cargo smoketest test_sql_format
cargo smoketest "cli::"
```

### Linting

```bash
# Full CI lint
cargo ci lint

# Individual
cargo fmt --all -- --check
cargo clippy --all --tests --benches -- -D warnings
```

### Benchmarks

```bash
cargo bench --bench generic --bench special
# Filter with regex:
cargo bench -- 'stdb_raw/.*/insert_bulk'
# Trigger on PR: comment "benchmarks please" or "callgrind please"
```

### TypeScript SDK

```bash
pnpm install && pnpm build && pnpm test
pnpm lint
```

### C# Bindings

```bash
dotnet test -warnaserror   # from crates/bindings-csharp
```

## Architecture

### Core Data Flow

```
Client (SDK) → WebSocket/HTTP → spacetimedb-client-api (Axum)
  → spacetimedb-core (module host controller)
    → Wasmtime/V8 runtime (executes WASM/JS modules)
      → spacetimedb-datastore (locking transactional DB)
        → spacetimedb-table (B-Tree physical storage)
    → spacetimedb-commitlog (WAL durability)
    → spacetimedb-subscription (push delta updates)
    → query pipeline (SQL execution)
```

### Key Crates

| Crate | Role |
|-------|------|
| `standalone` | Server binary entry point |
| `core` | Engine: module hosting, transaction scheduling, WAL, subscriptions, SQL |
| `datastore` | In-memory transactional relational tables with MVCC-style locking |
| `table` | Physical storage: B-Tree indexes, row scanning, page layout |
| `commitlog` | Write-ahead log for durability |
| `sats` | SpacetimeDB Algebraic Type System — core types + BSATN serialization |
| `lib` | Public types shared between modules and host |
| `client-api` | HTTP/WebSocket API server (Axum) |
| `client-api-messages` | WebSocket message schema (v1+v2); must stay in sync with all SDKs |
| `subscription` | Real-time subscription tracking and delta dispatch |
| `vm`, `query`, `expr`, `execution`, `physical-plan`, `sql-parser` | Query pipeline |
| `bindings` / `bindings-macro` / `bindings-sys` | Rust module SDK (proc macros + WASM FFI) |
| `cli` | The `spacetime` CLI tool |
| `codegen` | Generates Rust/C#/TypeScript/C++ client bindings from module schema |
| `pg` | PostgreSQL wire protocol compatibility (pgwire) |

### Module Runtimes

- **Wasmtime** — primary runtime for Rust/C#/C++ modules (`crates/core/src/host/wasmtime/`)
- **V8** — JavaScript/TypeScript module runtime (`crates/core/src/host/v8/`)

### Client SDKs

| SDK | Location |
|-----|----------|
| Rust | `sdks/rust/` |
| C# | `sdks/csharp/` |
| Unreal | `sdks/unreal/` |
| TypeScript | `crates/bindings-typescript/` |

### Module Bindings (server-side)

| Language | Location |
|----------|----------|
| Rust | `crates/bindings/`, `crates/bindings-macro/`, `crates/bindings-sys/` |
| C# | `crates/bindings-csharp/` |
| TypeScript | `crates/bindings-typescript/` |
| C++ | `crates/bindings-cpp/` |

## Key Design Decisions

- All application state is **in-memory**; durability via WAL (commitlog)
- **SATS** (SpacetimeDB Algebraic Type System) and **BSATN** binary serialization are the universal type/wire format across all languages
- Modules are WASM binaries — reducer panics/errors trigger full transaction rollback
- `clippy.toml` disallows `println!`, `print!`, `eprintln!`, `eprint!`, `dbg!` — use the `log` crate instead

## Toolchain

- **Rust:** pinned to `1.93.0` via `rust-toolchain.toml`; requires `wasm32-unknown-unknown` target
- **Node.js:** `>=18.0.0`, package manager `pnpm@9.7.0`
- **.NET:** version pinned via `global.json`
- **Cargo workspace:** 70+ members, `default-members` = CLI + standalone + update

## Smoketest Notes

Smoketests spawn real server processes and compile WASM modules (~15-20s per test). **Always use `cargo smoketest`** which rebuilds binaries first. Running `cargo test -p spacetimedb-smoketests` directly risks testing against stale binaries.

## PR Conventions

PRs require: description of changes, API/ABI breaking change labels, complexity rating (1-5), and a testing checklist. Comment "benchmarks please" on a PR to trigger CI benchmark runs.
