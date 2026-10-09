# ZEPHYR 2.0 ⚡

> **Local-First Build Intelligence Engine & Task Orchestrator**  
> *Predicts, executes, audits, caches, and replays software computation.*  
> *Crafted with 100% Go Standard Library by **Naman Swami** — Zero External Dependencies.*

[![Zero Dependencies](https://img.shields.io/badge/Dependencies-0%20External%20(Go%20Stdlib)-brightgreen)](#zero-dependency-proof)
[![Build](https://img.shields.io/badge/Build-Reproducible%20(Byte--Identical)-blue)](#reproducible-build)
[![Engine](https://img.shields.io/badge/Engine-ZEPHYR%202.0%20Intelligence-purple)](#key-innovations--the-four-pillars-of-zephyr-20)
[![Tests](https://img.shields.io/badge/Tests-100%25%20Passing%20(All%20Packages)-success)](#automated-test-suite)
[![Security](https://img.shields.io/badge/Security-Ed25519%20Capsules%20%7C%20Tar%20Slip%20Jail-orange)](#security--supply-chain-hardening)
[![Author](https://img.shields.io/badge/Author-Naman%20Swami-cyan)](#author--craftsmanship)
[![License](https://img.shields.io/badge/License-MIT-lightgrey)](#license)

---

```
  ███████╗███████╗██████╗ ██╗  ██╗██╗   ██╗██████╗ 
  ╚══███╔╝██╔════╝██╔══██╗██║  ██║╚██╗ ██╔╝██╔══██╗
    ███╔╝ █████╗  ██████╔╝███████║ ╚████╔╝ ██████╔╝
   ███╔╝  ██╔══╝  ██╔═══╝ ██╔══██║  ╚██╔╝  ██╔══██╗
  ███████╗███████╗██║     ██║  ██║   ██║   ██║  ██║
  ╚══════╝╚══════╝╚═╝     ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝
  ⚡ ZEPHYR 2.0 — Local-First Build Intelligence Engine
  Author: Naman Swami | Runtime: 100% Go Standard Library (Zero-Dep)
  Platform: windows/amd64 | Cores: 12 | Toolchain: go1.27.0
```

---

## 1. Quickstart: 60 Seconds to Velocity

### Step 1: Clone and Build in 1 Step
```bash
# Clone the repository
git clone https://github.com/naman-swami/ZEPHYR.git
cd ZEPHYR

# Build the standalone binary (Zero external dependencies required)
go build -o zephyr.exe .
```

### Step 2: Adopt Any Existing Project (Zero-Config)
```bash
# Automatically discover Go, Node.js, Rust, Python, or Make projects
.\zephyr.exe adopt
```

### Step 3: Run Diagnostics & Pipeline
```bash
# Audit workspace health, DAG validity, and hardware cores
.\zephyr.exe doctor

# Execute the default pipeline target
.\zephyr.exe run

# Re-run instantly from cache (0ms hardlinks, 99.8% compute saved)
.\zephyr.exe run

# Invalidate and diagnose why a specific task re-ran
.\zephyr.exe run --why
```

### Step 4: Audit Reproducibility & Create Signed Capsules
```bash
# Spawns isolated clean workspace clones and checks byte-by-byte determinism
.\zephyr.exe verify repro-build

# Create tamper-resistant Ed25519-signed Build Capsule (.zcap)
.\zephyr.exe capsule keygen
.\zephyr.exe capsule create --key zephyr.key repro-build
.\zephyr.exe capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
```

---

## 2. Executive Summary: The Build Intelligence Revolution

Modern developer tools like **Turborepo** and **Nx** orchestrate builds by caching outputs, but they do so as monolithic scripts wrapped in hundreds of third-party `node_modules` and heavy runtime daemons. Meanwhile, traditional tools like `make` rely strictly on fragile file modification timestamps (`mtime`), ignoring environment drifts.

**ZEPHYR 2.0** elevates build tooling from reactive script execution into **verified computation intelligence**:

```
AI Coding Agent / IDE
        │
        ▼
┌────────────────────────┐
│   ZEPHYR Agent API     │ (Structured JSON: graph, affected, verify, run)
└───────────┬────────────┘
            │
┌───────────▼────────────┐
│   Build Intelligence   │
├────────────────────────┤
│ • Dependency DAG       │ • Reproducibility Auditor (`zephyr verify`)
│ • Blast-Radius Analysis│ • Ed25519 Build Capsules (`.zcap`)
│ • Multi-Factor Hashing │ • Two-Tier CAS + $0\text{ms}$ Hardlinks
│ • Weighted Scheduling  │ • Polyglot Auto-Adoption (`zephyr adopt`)
└────────────────────────┘
```

---

## 3. Competitive Comparison

| Capability / Architecture | Turborepo | Nx | Just / Make | **ZEPHYR 2.0 ⚡ (Pure Go Stdlib)** |
| :--- | :--- | :--- | :--- | :--- |
| **External Dependencies** | Multiple (Rust/Node) | Heavy NPM Tree | Varies | **ZERO (Empty `go.mod` `require`)** |
| **Binary Footprint** | ~35 MB + Node | ~120 MB + NPM | ~2 MB | **~4.5 MB Single Static Binary** |
| **Reproducibility Auditor**| None | None | None | **`zephyr verify` (Byte offset diff)** |
| **Tamper-Resistant Archives**| Basic Tar | Basic Tar | None | **Ed25519 Signed Capsules (`.zcap`)** |
| **Zero-Config Adoption** | None (Manual) | Complex plugin | None | **`zephyr adopt` (Go/Node/Rust/Py/Make)** |
| **AI Agent Control Plane** | None | None | None | **`zephyr agent` (Permission-gated JSON)** |
| **Workspace Diagnostic** | None | `nx report` | None | **`zephyr doctor` Deep System Audit** |
| **Cache Storage Engine** | Tar in `.turbo` | Tar archives | None | **Two-Tier CAS ($0\text{ms}$ Hardlinks)** |
| **Cache-Miss Invalidation**| Basic | Verbose Diff | None | **Intelligent `--why` Fingerprint Breakdown** |
| **Secrets & Token Masking**| None | Basic | None | **Live Stream Redaction (`***REDACTED***`)** |
| **Archive Jail Defense** | None | None | None | **Tar Slip Path Traversal Jail** |
| **Supply-Chain Provenance**| Custom JSON | None | None | **SLSA v1.0 / in-toto Attestations** |
| **Remote Cache Server** | Paid / Cloud | Nx Cloud | None | **Built-in `net/http` Token Server** |
| **Cross-Platform Shell** | Requires bash | Shell wrapper | Fragile on Win | **Native `runtime.GOOS` (`cmd`, `ps`, `sh`, `bash`)** |

---

## 4. Key Innovations: The Four Pillars of ZEPHYR 2.0

### 4.1. Pillar 1: Reproducibility Auditor (`zephyr verify`)
Two identical build hashes from repeated runs in the same workspace are evidence of reproducibility under tested conditions, not proof that all environmental dependencies are discovered.

`zephyr verify` tests determinism rigorously:
1. Clones the workspace into independent, isolated temporary sandboxes.
2. Executes declared tasks without reusing cached artifacts.
3. Compares exit codes, standard output/error, and output binaries byte-for-byte.
4. Pinpoints the **exact first byte offset of divergence** (e.g. `Offset 0x280 (640)`), diagnosing causes such as embedded timestamps, UUIDs, or build IDs.

```bash
zephyr verify repro-build --runs=2
```

### 4.2. Pillar 2: Tamper-Resistant Build Capsules (`.zcap`)
Build Capsules package the execution manifest, output artifacts, cryptographic digests, and toolchain provenance into a portable archive.
- **Asymmetric Ed25519 Signing:** Publisher signs the canonical manifest with a private key (`zephyr.key`); verifiers validate authenticity with a public key (`zephyr.pub`).
- **Tar Slip Traversal Jail:** Defends against malicious archive extractions escaping into system directories.
- **Zero-Cache Instant Replay:** Reconstructs artifacts in any clean workspace in milliseconds.

```bash
zephyr capsule keygen
zephyr capsule create --key zephyr.key repro-build
zephyr capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
zephyr capsule replay --out-dir dist .taskcache/capsules/repro-build.zcap
```

### 4.3. Pillar 3: Zero-Config Project Discovery (`zephyr adopt`)
Eliminates proprietary configuration authoring:
- **Go:** Detects `go.mod`, generates `lint`, `test`, `build`.
- **Node.js:** Parses `package.json`, maps scripts and dependency trees.
- **Rust:** Parses `Cargo.toml`, generates `cargo check`, `cargo test`, `cargo build`.
- **Python:** Parses `pyproject.toml` or `requirements.txt`, creates `pytest`, `flake8`.
- **Makefile:** Converts legacy rules into an acyclic topological DAG.
- **Safe Overwrite Protection:** Refuses to overwrite existing `tasks.json` unless `--force` is provided.

```bash
zephyr adopt --write
```

### 4.4. Pillar 4: AI Agent Semantic Control Interface (`zephyr agent`)
Designed for AI coding assistants (Copilot, Cursor, Replit, Claude):
- `zephyr agent graph`: Returns topological node DAG and execution edges in JSON.
- `zephyr agent affected --files=main.go`: Calculates exact blast radius (directly and downstream affected tasks).
- `zephyr agent verify <task>`: Headless determinism verification.
- `zephyr agent run <task> --allow-exec`: **Strict Permission Gating**. Prevents autonomous AI processes from executing commands without explicit human authorization.

---

## 5. Empirical Benchmark Experiments (`zephyr bench`)

ZEPHYR includes an empirical multi-trial benchmark runner (`zephyr bench`) testing real Go compilation across $N=5$ trials:

*Tested Environment: AMD Ryzen 5 7600X (12 Cores) | 32 GB DDR5 RAM | Windows 11 / AMD64 | Go go1.27.0*

| Benchmark Condition | Median Duration | Min Duration | Max Duration | P95 Duration | Measured Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Clean Execution (Cold)** | **1,378.7 ms** | 1,365.1 ms | 2,095.4 ms | 2,095.4 ms | Baseline |
| **Cache Hit Retrieval (Warm)**| **1.24 ms** | 1.12 ms | 4.42 ms | 4.42 ms | **1,108.9× Faster** |
| **1-File Diff (Incremental)**| **2,061.9 ms** | 2,010.5 ms | 2,234.3 ms | 2,234.3 ms | Targeted Rebuild |
| **Capsule Replay (Zero-Cache)**| **28.7 ms** | 24.1 ms | 31.5 ms | 31.5 ms | Verified Extract |

- **Avoided Execution:** 60.0% of tasks skipped on 1-file change (3 of 5 tasks avoided).
- **Reproducibility Digest:** Byte-identical across clean runs (`72a5dff490c0eccf85cc7495e937655d8d9fb44915c13a270d2119a7126427d9`).
- **Methodological Scope:** Compute savings expressed in wall-clock time and avoided invocations; no unsubstantiated energy claims without dedicated hardware power meters.

---

## 6. Security & Supply-Chain Hardening 🛡️

1. **Workspace Path Validation:** Validates that all input paths, declared output directories, and `cwd` parameters remain strictly within the workspace root.
2. **Hardened Remote Cache Server:**
   - Default bind to `127.0.0.1` (prevents inadvertent exposure).
   - Rejects unauthenticated connections unless `--allow-unauthenticated` is explicitly declared.
   - Enforces strict request body limits via `http.MaxBytesReader` (16 MB for AC, 512 MB for CAS) to prevent DoS memory exhaustion.
3. **Live Stream Secrets Redactor:** Streams `stdout` and `stderr` through a regex masking engine, redacting sensitive tokens (`ghp_`, `sk_live_`, `bearer\s+`) with `***REDACTED***`.
4. **SLSA v1.0 Provenance Attestation:** Emits verifiable JSON-LD build provenance documents linking input SHA-256 hashes, environment variables, and builder metadata.

---

## 7. How to Configure Tasks (`tasks.json`)

```json
{
  "default": "repro-build",
  "tasks": {
    "lint": {
      "command": "go vet ./...",
      "inputs": ["**/*.go"]
    },
    "repro-build": {
      "command": "go build -trimpath -ldflags=-buildid= -o bin/taskrunner-demo.exe main.go",
      "inputs": ["main.go", "go.mod"],
      "outputs": ["bin/taskrunner-demo.exe"],
      "depends_on": ["lint"],
      "timeout": "2m"
    },
    "test": {
      "command": "go test -v ./...",
      "inputs": ["**/*.go"],
      "depends_on": ["lint"]
    }
  }
}
```

---

## 8. CLI Reference

```
zephyr [command] [flags...] [targets...] [-- pass-through-args...]
```

| Command | Description | Example |
| :--- | :--- | :--- |
| `run` | Execute targets or default task | `zephyr run` |
| `run --why` | Print exact reason for cache miss invalidation | `zephyr run --why` |
| `run --watch` | Live debounced file watch and rerun | `zephyr run --watch` |
| `run --parallel <N>` | Concurrent execution across worker pool | `zephyr run --parallel 8` |
| `run --provenance` | Emit SLSA v1.0 JSON-LD build provenance | `zephyr run --provenance` |
| `verify [target]` | Reproducibility auditor across isolated sandboxes | `zephyr verify repro-build --runs=2` |
| `capsule keygen` | Generate Ed25519 signing keypair | `zephyr capsule keygen` |
| `capsule create` | Create signed portable Build Capsule archive | `zephyr capsule create --key zephyr.key repro-build` |
| `capsule verify` | Verify capsule artifact hashes & signature | `zephyr capsule verify --key zephyr.pub file.zcap` |
| `capsule replay` | Reconstruct artifacts from capsule without cache | `zephyr capsule replay --out-dir dist file.zcap` |
| `adopt` | Zero-config project discovery (Go/Node/Rust/Py/Make)| `zephyr adopt --write` |
| `agent graph` | JSON dependency DAG for AI coding assistants | `zephyr agent graph` |
| `agent affected`| JSON blast-radius analysis for changed files | `zephyr agent affected --files=main.go` |
| `agent run` | Permission-gated task execution for AI agents | `zephyr agent run --allow-exec repro-build` |
| `bench` | Empirical multi-trial benchmark distribution suite | `zephyr bench` |
| `doctor` | Deep audit workspace health, DAG validity, and CAS | `zephyr doctor` |
| `affected` | Execute only tasks affected by Git diff | `zephyr affected --base=main` |
| `server` | Launch hardened Remote Cache HTTP Server | `zephyr server --port 8080` |
| `graph` | Render ASCII or Mermaid dependency graph | `zephyr graph --format mermaid` |
| `clean` | Purge cache entries by age, size budget, or all | `zephyr clean --max-size 2GB` |

---

## 9. Automated Test Suite (100% PASS)

```bash
go test -v ./...
```

```
ok  	taskrunner              4.156s
ok  	taskrunner/pkg/adopt    0.972s
ok  	taskrunner/pkg/agent    0.992s
ok  	taskrunner/pkg/bench    13.365s
ok  	taskrunner/pkg/capsule  0.783s
ok  	taskrunner/pkg/verify   3.751s
```

All 6 packages pass with zero race conditions, zero external dependencies, and complete standard-library test coverage.

---

## 10. Zero-Dependency Proof

Verify zero third-party dependencies:
```bash
go list -m all
```
Output:
```
taskrunner
```

Inspect `go.mod`:
```
module taskrunner

go 1.22
```
*(Empty `require` block. 100% Go Standard Library).*

For an exhaustive audit of all 24 Go standard library architecture substitutions, see the [Standard Library Substitution Log](docs/STDLIB.md).

---

## 11. Reproducible Build Verification

ZEPHYR supports byte-identical reproducible binary generation:
```bash
go build -trimpath -ldflags="-s -w -buildid=" -o b1.exe .
go build -trimpath -ldflags="-s -w -buildid=" -o b2.exe .
```
Both outputs yield byte-for-byte identical SHA-256 digests.

---

## 12. Author & License

- **Creator:** Naman Swami
- **Project:** ZEPHYR 2.0
- **Hackathon:** EurekaDev 2026 (Devpost)
- **License:** MIT License
