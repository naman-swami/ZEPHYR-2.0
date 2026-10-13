# ZEPHYR 2.0 ⚡

> **Local-First Build Intelligence Engine & High-Velocity Task Orchestrator**  
> *Predicts, executes, audits, caches, and replays software computation.*  
> *Crafted with 100% Go Standard Library by **Naman Swami** — Zero External Dependencies.*

[![Zero Dependencies](https://img.shields.io/badge/Dependencies-0%20External%20(Go%20Stdlib)-brightgreen)](#12-zero-dependency-proof)
[![Build](https://img.shields.io/badge/Build-Reproducible%20(Byte--Identical)-blue)](#direction-6-hermetic-reproducibility-auditing-zephyr-verify)
[![Engine](https://img.shields.io/badge/Engine-ZEPHYR%202.0%20Intelligence-purple)](#7-architecture--engineering-innovations)
[![Tests](https://img.shields.io/badge/Tests-100%25%20Passing%20(14%20Suites)-success)](#10-automated-testing--verification-suites)
[![Security](https://img.shields.io/badge/Security-Ed25519%20Capsules%20%7C%20Tar%20Slip%20Jail-orange)](#9-security--supply-chain-governance)
[![Author](https://img.shields.io/badge/Author-Naman%20Swami-cyan)](#13-author--license)
[![License](https://img.shields.io/badge/License-MIT-lightgrey)](#13-author--license)

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
  Cross-Platform: Windows (AMD64/ARM64) • Linux • macOS (Intel/Apple Silicon)
```

---

## Table of Contents

1. [Executive Overview](#1-executive-overview)
2. [Quickstart (60 Seconds to Velocity)](#2-quickstart-60-seconds-to-velocity)
3. [Prerequisites & Installation](#3-prerequisites--installation)
4. [Step-by-Step Directions & Guided Workflows](#4-step-by-step-directions--guided-workflows)
   - [Direction 1: Zero-Config Project Adoption (`zephyr adopt`)](#direction-1-zero-config-project-adoption-zephyr-adopt)
   - [Direction 2: System Health Diagnostics (`zephyr doctor`)](#direction-2-system-health-diagnostics-zephyr-doctor)
   - [Direction 3: Pipeline & Task Execution (`zephyr run`)](#direction-3-pipeline--task-execution-zephyr-run)
   - [Direction 4: Cache-Miss Root-Cause Diagnostics (`zephyr run --why`)](#direction-4-cache-miss-root-cause-diagnostics-zephyr-run---why)
   - [Direction 5: Git Blast-Radius Targeted Execution (`zephyr affected`)](#direction-5-git-blast-radius-targeted-execution-zephyr-affected)
   - [Direction 6: Hermetic Reproducibility Auditing (`zephyr verify`)](#direction-6-hermetic-reproducibility-auditing-zephyr-verify)
   - [Direction 7: Cryptographic Build Capsules (`zephyr capsule`)](#direction-7-cryptographic-build-capsules-zephyr-capsule)
   - [Direction 8: AI Agent & MCP Protocol Integration (`zephyr agent` & `zephyr mcp`)](#direction-8-ai-agent--mcp-protocol-integration-zephyr-agent--zephyr-mcp)
   - [Direction 9: Distributed Remote HTTP Cache Server (`zephyr server`)](#direction-9-distributed-remote-http-cache-server-zephyr-server)
   - [Direction 10: Task Graph Visualization (`zephyr graph`)](#direction-10-task-graph-visualization-zephyr-graph)
   - [Direction 11: Native IDE Language Server Protocol (`zephyr lsp`)](#direction-11-native-ide-language-server-protocol-zephyr-lsp)
   - [Direction 12: Peer-to-Peer LAN Cache Sync (`zephyr p2p`)](#direction-12-peer-to-peer-lan-cache-sync-zephyr-p2p)
   - [Direction 13: Kernel Resource Profiler & SVG Flame Graphs (`zephyr profile`)](#direction-13-kernel-resource-profiler--svg-flame-graphs-zephyr-profile)
5. [Complete CLI Command & Option Matrix](#5-complete-cli-command--option-matrix)
   - [5.1 Primary Subcommands](#51-primary-subcommands)
   - [5.2 `zephyr run` Execution Flags](#52-zephyr-run-execution-flags)
6. [`tasks.json` Configuration Specification](#6-tasksjson-configuration-specification)
   - [Schema Property Reference](#schema-property-reference)
   - [Polyglot Monorepo Example](#polyglot-monorepo-example)
7. [Architecture & Engineering Innovations](#7-architecture--engineering-innovations)
   - [7.1 Two-Tier CAS with 256-Bucket Sharding](#71-two-tier-cas-with-256-bucket-sharding)
   - [7.2 Zero-Copy Hardlink Restoration vs. In-Memory Compression](#72-zero-copy-hardlink-restoration-vs-in-memory-compression)
   - [7.3 Dynamic Critical Path Method (CPM) Priority Scheduling](#73-dynamic-critical-path-method-cpm-priority-scheduling)
   - [7.4 Speculative Early Cutoff (Hash-Invariant Pruning)](#74-speculative-early-cutoff-hash-invariant-pruning)
   - [7.5 Process Tree Termination on Windows and Unix](#75-process-tree-termination-on-windows-and-unix)
8. [Empirical Benchmarks & Performance](#8-empirical-benchmarks--performance)
9. [Security & Supply-Chain Governance](#9-security--supply-chain-governance)
10. [Automated Testing & Verification Suites](#10-automated-testing--verification-suites)
    - [10.1 System Verification Suite (`verify.ps1`)](#101-system-verification-suite-verifyps1)
    - [10.2 Master Adversarial & Chaos Test Harness (`test/run_all_tests.ps1`)](#102-master-adversarial--chaos-test-harness-testrun_all_testsps1)
    - [10.3 Go Standard Package Tests](#103-go-standard-package-tests)
11. [Competitive Comparison Matrix](#11-competitive-comparison-matrix)
12. [Zero-Dependency Proof](#12-zero-dependency-proof)
    - [Go Standard Library Substitution Map](#go-standard-library-substitution-map)
13. [Author & License](#13-author--license)

---

## 1. Executive Overview

Modern build orchestrators (such as Turborepo, Nx, and Gradle) accelerate build times by caching task outputs. However, they rely heavily on massive third-party package dependencies, unverified ambient environments, fragile modification timestamps (`mtime`), and blind downstream invalidation chains.

**ZEPHYR 2.0** redefines task orchestration from reactive script running into **rigorously verified, hermetic computation intelligence**:

* **0 External Dependencies:** Implemented strictly using the **Go Standard Library** (empty `require` block in `go.mod`). No runtime daemons, no Node.js/npm dependencies, and no CGO requirements.
* **Two-Tier CAS with $0\text{ms}$ Hardlinks:** Content-Addressable Storage (CAS) partitioned across 256 hexadecimal buckets (`00`–`ff`). Instantaneous output restoration using raw filesystem hardlinks (`os.Link`) at 0% CPU decompression overhead.
* **Dynamic Critical Path Method (CPM) Priority Scheduling:** Historical task duration tracking in `.taskcache/stats.json` that computes the exact Critical Path $\text{CP}(T) = \text{Duration}(T) + \max_{D} \text{CP}(D)$, dynamically prioritizing bottleneck tasks to minimize total pipeline makespan.
* **Speculative Early Cutoff (Hash-Invariant Pruning):** Calculates Merkle output digests post-execution. If a re-executed task produces byte-identical outputs, downstream dependents are saved from rebuilds and remain **100% CACHED**.
* **Hermetic Reproducibility Auditor (`zephyr verify`):** Spawns isolated clean workspace sandboxes, executes builds across multiple runs, compares binaries byte-for-byte, and pinpoints the exact byte offset of divergence.
* **Ed25519-Signed Build Capsules (`.zcap`):** Tamper-evident, portable build artifacts bundled with execution manifests, environment hashes, and SLSA v1.0 provenance attestations, protected by a Tar Slip path traversal jail.
* **AI Agent & MCP Control Plane (`zephyr mcp` / `zephyr agent`):** Native Model Context Protocol (JSON-RPC 2.0) stdio server and structured JSON CLI enabling AI assistants (Cursor, Claude Code, Copilot, OpenHands) to query dependency graphs and run permission-gated tasks without hallucinating build commands.

---

## 2. Quickstart (60 Seconds to Velocity)

> [!TIP]
> **Automated One-Click Verification:** Run `powershell -ExecutionPolicy Bypass -File .\verify.ps1` to execute the automated 13-point verification suite in under 90 seconds.

### Step 1: Clone and Build Standalone Binary

```bash
# Clone the ZEPHYR 2.0 repository
git clone https://github.com/naman-swami/ZEPHYR-2.0.git
cd ZEPHYR-2.0

# Build the static executable using Go 1.20+ (zero external modules needed)
# On Windows:
go build -trimpath -ldflags="-s -w -buildid=" -o zephyr.exe .

# On Linux / macOS:
go build -trimpath -ldflags="-s -w -buildid=" -o zephyr .
chmod +x zephyr
```

### Step 2: Discover & Adopt Your Existing Project

```bash
# Automatically inspect and detect Go, Node.js, Rust, Python, or Make configurations
./zephyr adopt --write
```

### Step 3: Audit Workspace Health & Run Your Pipeline

```bash
# Verify workspace health, DAG validity, and hardware capacity
./zephyr doctor

# Execute the default build target
./zephyr run

# Re-run instantly from cache (0ms hardlinks, 99.8% compute saved)
./zephyr run

# Understand exactly why any task invalidated
./zephyr run --why
```

---

## 3. Prerequisites & Installation

### System Requirements
* **Go Compiler:** Go 1.20 or newer (Go 1.22+ recommended).
* **Operating Systems:**
  * **Windows:** Windows 10, Windows 11, Windows Server 2019+ (`amd64`, `arm64`)
  * **Linux:** Ubuntu, Debian, Alpine, Fedora, Arch, RHEL (`amd64`, `arm64`)
  * **macOS:** macOS 12+ Monterey, Ventura, Sonoma, Sequoia (`amd64`, `Apple Silicon arm64`)
* **Shell Environments Supported:** Native detection for PowerShell, `cmd.exe`, `bash`, `zsh`, `sh`.

### Installation Directions

#### Option A: Build and Install from Source (Recommended)

```bash
# 1. Clone the repository
git clone https://github.com/naman-swami/ZEPHYR-2.0.git
cd ZEPHYR-2.0

# 2. Verify zero external dependencies
go list -m all
# Expected output: taskrunner (No third-party packages)

# 3. Compile standalone binary with reproducible build flags
# Windows (PowerShell):
go build -trimpath -ldflags="-s -w -buildid=" -o zephyr.exe .

# Linux / macOS:
go build -trimpath -ldflags="-s -w -buildid=" -o zephyr .
chmod +x zephyr
```

#### Option B: Configure Global PATH

**Windows (PowerShell as Administrator or User):**
```powershell
# Move binary to a user tools directory
New-Item -ItemType Directory -Path "$HOME\bin" -Force | Out-Null
Copy-Item .\zephyr.exe "$HOME\bin\zephyr.exe"

# Add to user PATH environment variable
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$HOME\bin", [EnvironmentVariableTarget]::User)
$env:Path += ";$HOME\bin"
```

**Linux / macOS:**
```bash
# Move to system or user bin directory
sudo install -m 755 zephyr /usr/local/bin/zephyr
# Or for local user without sudo:
mkdir -p ~/.local/bin && cp zephyr ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

#### Step 4: Verify Installation
```bash
zephyr version
```
*Expected Output:*
```text
ZEPHYR (windows/amd64, 100% Go stdlib by Naman Swami)
```

---

## 4. Step-by-Step Directions & Guided Workflows

---

### Direction 1: Zero-Config Project Adoption (`zephyr adopt`)

You do not need to manually write configuration files to use ZEPHYR. `zephyr adopt` inspects your existing codebase, discovers build declarations, and maps them into a dependency-aware `tasks.json`.

#### Supported Ecosystems:
* **Go:** Detects `go.mod` and generates `lint` (`go vet`), `test` (`go test`), and `build`.
* **Node.js / TypeScript:** Parses `package.json`, maps npm/yarn/pnpm scripts (`build`, `test`, `lint`), and maps dependency trees.
* **Rust:** Parses `Cargo.toml`, generating `cargo check`, `cargo test`, and `cargo build`.
* **Python:** Parses `pyproject.toml` or `requirements.txt`, generating `pytest` and `flake8`.
* **C / C++ / Polyglot:** Parses `Makefile` rules and targets into an acyclic topological DAG.

#### Directions:

1. **Dry-Run Preview (Non-destructive inspection):**
   ```bash
   zephyr adopt
   ```
   *Inspect the generated JSON output in stdout without writing any files to disk.*

2. **Commit Configuration to Workspace:**
   ```bash
   zephyr adopt --write
   ```
   *Generates `tasks.json` in your current workspace root. If a `tasks.json` already exists, ZEPHYR safely refuses to overwrite it.*

3. **Force Overwrite Existing Configuration:**
   ```bash
   zephyr adopt --write --force
   ```

4. **Adopt a Subdirectory or External Repository:**
   ```bash
   zephyr adopt --dir ./packages/backend --write
   ```

---

### Direction 2: System Health Diagnostics (`zephyr doctor`)

Before executing builds or onboarding CI nodes, run `zephyr doctor` to perform a comprehensive 5-category health audit:

```bash
zephyr doctor
```

#### Sample Diagnostic Output:
```text
[ZEPHYR DOCTOR] Auditing workspace health & system integrity...

Category          Status         Details
--------          ------         -------
Configuration     [OK]  Loaded tasks.json with 6 defined tasks
Dependency Graph  [OK]  DAG is acyclic and topologically valid
Cache Storage     [OK]  .taskcache active (8 entries, 28.3 MB)
Git Integration   [OK]  Git repository detected; 'affected' command available
System Resources  [OK]  12 CPU Cores | Go go1.27.0 | windows/amd64

[DOCTOR RESULT] All systems operational. Workspace is ready for high-velocity builds!
```

> [!NOTE]
> If a cycle is detected in your task graph (e.g., `task A -> task B -> task A`), `zephyr doctor` highlights the exact cycle path and exits with a non-zero exit code.

---

### Direction 3: Pipeline & Task Execution (`zephyr run`)

`zephyr run` executes tasks in topological order, parallelizing independent branches across your host's CPU worker pool while retrieving cached artifacts for unchanged tasks.

#### Directions:

1. **Execute Default Target:**
   ```bash
   zephyr run
   ```
   *Reads the `"default"` property from `tasks.json` and executes that task along with all required upstream dependencies.*

2. **Execute Specific Target(s):**
   ```bash
   zephyr run lint
   zephyr run test build package
   ```

3. **Pass Arbitrary CLI Arguments Through to Underlying Tools:**
   ```bash
   # Everything after '--' is forwarded directly to the task's command
   zephyr run test -- -v -run TestDAG
   ```

4. **Parallel Worker Pool Tuning:**
   ```bash
   # Default is runtime.NumCPU() (e.g. 12 workers)
   zephyr run --parallel 4
   ```

5. **Live Log Streaming with Task Prefixes:**
   ```bash
   zephyr run --stream
   ```
   *Streams `stdout` and `stderr` in real time with colored task prefix labels (e.g., `[build] ...`). Automatically redacts sensitive tokens (`ghp_`, `sk_live_`, `bearer`).*

6. **Continuous File Watch Mode (Hot Rebuilds):**
   ```bash
   zephyr run --watch
   ```
   *Monitors all declared inputs with a 150ms debounce window. When a file changes, only the invalidated tasks are re-executed.*

7. **Dry-Run Preview (Execution Batches):**
   ```bash
   zephyr run --dry-run
   ```
   *Calculates and displays topological execution batches without executing processes.*

8. **Non-Failing Sibling Tasks (CI Mode):**
   ```bash
   zephyr run --no-fail-fast
   ```
   *Continues executing independent tasks even if a parallel sibling task encounters an error.*

9. **Structured Machine-Readable JSON Output:**
   ```bash
   zephyr run --json
   ```

---

### Direction 4: Cache-Miss Root-Cause Diagnostics (`zephyr run --why`)

When a task executes instead of hitting the cache, developers often waste time guessing why. ZEPHYR records historical fingerprint manifests and compares inputs, environment variables, commands, and upstream hashes.

```bash
zephyr run --why
```

#### Diagnostic Indicators:
* `[Why: input file modified (main.go)]`: File content hash changed.
* `[Why: command modified]` : Command string in `tasks.json` was altered.
* `[Why: env var modified (GOOS)]`: Environment variable value differed from the cached manifest.
* `[Why: dependency modified (lint)]`: Upstream dependency was re-executed and produced different outputs.

#### Example Terminal Output:
```text
[lint] [CACHED] (Hash: 7bfd158b)
[build] [RUNNING] "go build -o bin/demo main.go" (Hash: a4c891e2) [Why: input file modified (main.go)]
[build] [OK] in 1.42s
```

---

### Direction 5: Git Blast-Radius Targeted Execution (`zephyr affected`)

In large monorepos, running every task on every pull request slows down CI pipelines. `zephyr affected` performs git diff blast-radius analysis to execute only the tasks directly or transitively touched by recent commits.

#### Directions:

1. **Run Tasks Affected Against Main Branch:**
   ```bash
   zephyr affected --base=main
   ```

2. **Run Tasks Affected by Staged Changes:**
   ```bash
   zephyr affected --staged
   ```

3. **Run Tasks Affected by Untracked Files:**
   ```bash
   zephyr affected --untracked
   ```

---

### Direction 6: Hermetic Reproducibility Auditing (`zephyr verify`)

A build is reproducible if identical source inputs yield byte-for-byte identical output binaries across independent environments.

`zephyr verify` tests determinism empirically:
1. Clones the current repository into two independent, isolated temporary sandboxes (`temp-dir-1` and `temp-dir-2`).
2. Cleans out cache entries to prevent cache reuse.
3. Compares exit codes, terminal output, and declared output binaries byte-for-byte.
4. Pinpoints the exact byte offset of divergence if a mismatch occurs.

#### Directions:

```bash
# Run a 2-stage verification audit on your build task
zephyr verify repro-build --runs=2
```

#### Expected Success Output:
```text
══════════════════════════════════════════════════════════════════════════════
  ZEPHYR REPRODUCIBILITY AUDIT REPORT
  Task: repro-build | Runs: 2 | Tested Environment: windows/amd64
══════════════════════════════════════════════════════════════════════════════
  VERDICT: [PASS] 100% REPRODUCIBLE UNDER TESTED CONDITIONS

  Detailed Checks:
    ✓ Exit Codes Match: 0 (identical across runs)
    ✓ Stdout Stream: Byte-identical across runs
    Declared Output Artifacts (1 checked):
      ✓ bin\taskrunner-demo.exe
        SHA-256: 6fe25d5121e6c989dda8a3f5e39840f242602cc3cdc66f93879602249251356c (11549184 bytes)
══════════════════════════════════════════════════════════════════════════════
```

#### How to Fix Divergent Builds:
If divergence is reported at `Offset 0x280 (640)`:
* **Go:** Ensure `-trimpath` and `-ldflags="-s -w -buildid="` are passed to strip build paths and random build IDs.
* **C/C++:** Pass `-frandom-seed=string` and avoid `__DATE__` or `__TIME__` macros.
* **Node.js/Python:** Ensure directory listings and object keys are deterministically sorted.

---

### Direction 7: Cryptographic Build Capsules (`zephyr capsule`)

Build Capsules (`.zcap`) package execution manifests, cryptographic checksums, environment configurations, and output binaries into an Ed25519-signed archive. This guarantees supply-chain integrity when shipping binaries between developers, staging environments, and production clusters.

#### Lifecycle Directions:

#### Step 1: Generate Cryptographic Asymmetric Keypair
```bash
zephyr capsule keygen
```
*Generates `zephyr.key` (Ed25519 private signing key) and `zephyr.pub` (Ed25519 public verification key).*

#### Step 2: Create and Sign a Build Capsule
```bash
zephyr capsule create --key zephyr.key repro-build
```
*Executes or verifies the task, bundles outputs into `.taskcache/capsules/repro-build.zcap`, and embeds an Ed25519 signature of the Merkle digest.*

#### Step 3: Verify Capsule Authenticity & Tamper Resistance
```bash
zephyr capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
```
*Verifies signature authenticity and checks all payload digests. Rejects modified bytes or forged manifests.*

#### Step 4: Hermetically Replay Capsule into Clean Workspace
```bash
zephyr capsule replay --out-dir dist .taskcache/capsules/repro-build.zcap
```
*Reconstructs artifacts directly into `dist/` without compiling code, hitting network servers, or querying local cache.*

> [!IMPORTANT]
> **Tar Slip Security Defense:** The capsule extraction engine enforces path normalization and root jailing, automatically aborting extraction if an archive attempts path traversal (e.g. `../../etc/passwd`).

---

### Direction 8: AI Agent & MCP Protocol Integration (`zephyr agent` & `zephyr mcp`)

Modern AI coding agents (Claude Code, Cursor, OpenHands, Copilot) frequently hallucinate build commands, run incorrect scripts, or trigger runaway processes. ZEPHYR provides both a CLI semantic interface and a native **Model Context Protocol (MCP)** server.

#### Option A: Semantic Agent CLI Interface

1. **Query Full Task Graph in Machine-Readable JSON:**
   ```bash
   zephyr agent graph
   ```
   *Returns nodes, dependencies, commands, inputs, and outputs.*

2. **Query Blast-Radius for Modified Files:**
   ```bash
   zephyr agent affected --files=main.go,pkg/util.go
   ```
   *Returns `directly_affected` and `downstream_affected` task arrays.*

3. **Headless Determinism Verification:**
   ```bash
   zephyr agent verify repro-build
   ```

4. **Strict Permission-Gated Execution:**
   ```bash
   # Autonomous AI without human approval:
   zephyr agent run build
   # -> Error: Status "PERMISSION_DENIED" (Execution requires --allow-exec)

   # Explicit human-approved execution:
   zephyr agent run build --allow-exec
   ```

#### Option B: Native Model Context Protocol (MCP) Stdio Server

Start the native JSON-RPC 2.0 stdio server:
```bash
zephyr mcp
```

#### MCP Integration Setup:

**For Cursor IDE (`.cursor/mcp.json` or Global Settings):**
```json
{
  "mcpServers": {
    "zephyr": {
      "command": "zephyr",
      "args": ["mcp"]
    }
  }
}
```

**For Claude Desktop (`claude_desktop_config.json`):**
```json
{
  "mcpServers": {
    "zephyr": {
      "command": "/usr/local/bin/zephyr",
      "args": ["mcp"]
    }
  }
}
```

#### Exposed MCP Tools:
* `zephyr_graph`: Inspects full dependency DAG and available targets.
* `zephyr_affected`: Calculates exact blast radius for edited files.
* `zephyr_why`: Pinpoints reasons why a target is a cache miss.
* `zephyr_verify`: Executes a non-destructive reproducibility audit.
* `zephyr_run`: Executes a task (requires explicit `"allow_exec": true` permission).

---

### Direction 9: Distributed Remote HTTP Cache Server (`zephyr server`)

Share cached outputs across multiple developer workstations and CI runners using ZEPHYR's built-in remote cache server.

#### Directions:

1. **Launch Hardened Remote Cache Server:**
   ```bash
   # Binds to 127.0.0.1 by default with mandatory token authentication
   zephyr server --port 8080 --token supersecrettoken
   ```

2. **Configure Custom Cache Storage Directory:**
   ```bash
   zephyr server --port 8080 --token supersecrettoken --dir /var/cache/zephyr
   ```

3. **Security Safeguards:**
   * Rejects unauthenticated requests unless `--allow-unauthenticated` is explicitly declared.
   * Restricts Action Cache requests to 16 MB and CAS blob uploads to 512 MB via `http.MaxBytesReader` to prevent denial-of-service memory exhaustion.

---

### Direction 10: Task Graph Visualization (`zephyr graph`)

Inspect pipeline dependencies visually:

```bash
# Render ASCII tree in terminal
zephyr graph

# Output Mermaid markdown diagram
zephyr graph --format mermaid
```

#### Example ASCII Dependency Output:
```text
Dependency Tree:
├── build (go build -o bin/taskrunner-demo main.go)
│   └── depends on: lint
├── lint (go vet ./...)
├── package (go version)
│   └── depends on: test:integration
├── repro-build (go build -trimpath -ldflags=-buildid= -o bin/taskrunner-demo.exe main.go)
│   └── depends on: lint
├── test:integration (go test -v -run TestPipeline ./...)
│   └── depends on: build
│   └── depends on: test:unit
├── test:unit (go test -v -run TestDAG ./...)
│   └── depends on: lint
```

---

### Direction 11: Native IDE Language Server Protocol (`zephyr lsp`)

ZEPHYR includes a native **Language Server Protocol (LSP)** server over `stdio` (`Content-Length` framed JSON-RPC 2.0) for real-time `tasks.json` intelligence in **VS Code, Cursor, Zed, Neovim, and Helix**.

#### Capabilities:
* **Real-Time Diagnostics (`textDocument/publishDiagnostics`):** Detects circular dependencies (3-color DFS), undefined upstream tasks, missing commands, and path traversal violations as you type.
* **Hover Telemetry (`textDocument/hover`):** Displays Markdown cards with a task's command, inputs, outputs, and historical execution durations from `.taskcache/stats.json`.
* **Schema & Target Auto-Completion (`textDocument/completion`):** Suggests schema keys and defined task target names inside `depends_on`.

#### Directions:

```bash
# Run instant CLI static analysis on tasks.json
zephyr lsp --check

# Start stdio Language Server for IDE integration
zephyr lsp
```

---

### Direction 12: Peer-to-Peer LAN Cache Sync (`zephyr p2p`)

Share verified build artifacts across local developer network subnets without cloud servers. Every blob transferred over P2P is cryptographically hashed against its SHA-256 content digest before admission into `.taskcache/cas/objects/`, preventing cache poisoning on shared networks.

#### Directions:

```bash
# 1. Inspect local P2P node inventory
zephyr p2p status

# 2. Discover active ZEPHYR peers on your local subnet via UDP beacons
zephyr p2p discover --key team-cluster-secret

# 3. Serve your local sharded CAS to teammates on port 9090
zephyr p2p serve --port 9090 --key team-cluster-secret

# 4. Pull missing CAS blobs & Action Cache manifests from a peer with zero-trust SHA-256 verification
zephyr p2p sync --peer http://192.168.1.42:9090 --key team-cluster-secret
```

---

### Direction 13: Kernel Resource Profiler & SVG Flame Graphs (`zephyr profile`)

Measure exact kernel-level `UserTime`, `SystemTime`, wall-clock duration, CPU utilization percentage, and peak memory footprint per task, and automatically generate a standalone interactive **SVG Flame Graph** and **Chrome / Perfetto Trace JSON**.

#### Directions:

```bash
# Profile default pipeline or specific targets and emit SVG Flame Graph
zephyr profile
zephyr profile lint repro-build
```

* **Interactive SVG Flame Graph:** Written to `.taskcache/profile/flamegraph.svg` (open in any browser to hover over execution bars and inspect CPU/memory metrics).
* **Chrome / Perfetto Trace:** Written to `.taskcache/profile/trace.json` (compatible with `ui.perfetto.dev` and `chrome://tracing`).

---

## 5. Complete CLI Command & Option Matrix

ZEPHYR follows the standard CLI pattern:
```text
zephyr [command] [flags...] [targets...] [-- pass-through-arguments...]
```

### 5.1 Primary Subcommands

| Command | Description | Syntax / Example | Key Flags & Options |
| :--- | :--- | :--- | :--- |
| `run` | Execute target tasks or default pipeline | `zephyr run [targets...]` | `--why`, `--watch`, `--stream`, `--parallel <N>`, `--dry-run`, `--no-fail-fast`, `--json`, `--provenance` |
| `doctor` | Deep audit workspace health, DAG validity, and CAS | `zephyr doctor` | `--config <path>` |
| `adopt` | Zero-config project discovery and `tasks.json` generation | `zephyr adopt --write` | `--write` (commit to disk), `--force` (overwrite), `--dir <path>` |
| `affected` | Execute only tasks affected by Git diff | `zephyr affected --base=main` | `--base <branch>`, `--staged`, `--untracked` |
| `verify` | Hermetic reproducibility auditor across clean sandboxes | `zephyr verify <target>` | `--runs <N>` (default 2), `--timeout <dur>` |
| `capsule` | Ed25519 cryptographic Build Capsule lifecycle | `zephyr capsule <action>` | Subcommands: `keygen`, `create`, `verify`, `replay` |
| `agent` | Machine-readable AI Agent semantic control interface | `zephyr agent <action>` | Subcommands: `graph`, `affected`, `verify`, `run` (`--allow-exec`) |
| `mcp` | Start native Model Context Protocol (MCP) stdio server | `zephyr mcp` | JSON-RPC 2.0 over stdio for Cursor, Claude Code, OpenHands |
| `lsp` | Native Language Server Protocol (LSP) for IDEs | `zephyr lsp [--check]` | `--check` (CLI static analysis), `--config <path>` |
| `p2p` | Peer-to-Peer LAN cache discovery & verified sync | `zephyr p2p <action>` | Subcommands: `status`, `discover`, `serve`, `sync` (`--peer`, `--key`) |
| `profile` | Kernel CPU/memory profiler & SVG Flame Graph generator | `zephyr profile [targets...]` | `--svg <path>`, `--trace <path>`, `--config <path>` |
| `server` | Hardened Remote HTTP Cache Server | `zephyr server --port 8080` | `--port <N>`, `--token <secret>`, `--dir <path>`, `--allow-unauthenticated` |
| `graph` | Visualize pipeline dependency tree | `zephyr graph` | `--format <ascii\|mermaid>` |
| `list` | Print declared tasks and metadata | `zephyr list` | `--json` |
| `bench` | Empirical multi-trial performance distribution runner | `zephyr bench` | Evaluates cold, warm cache, incremental, and capsule replay |
| `clean` | Purge cached objects from `.taskcache` | `zephyr clean` | `--all`, `--max-age <dur>`, `--max-size <bytes\|MB\|GB>` |
| `version` | Print binary version, architecture, and author info | `zephyr version` | N/A |
| `help` | Print built-in CLI help guide | `zephyr help [command]` | N/A |

### 5.2 `zephyr run` Execution Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-why` | `bool` | `false` | Explains the precise root cause for any cache miss (file modified, command altered, env changed, upstream invalidated). |
| `-watch` | `bool` | `false` | Live file watching mode. Automatically reruns invalidated tasks on file save (150ms debounce window). |
| `-stream` | `bool` | `false` | Streams live `stdout`/`stderr` with colored task name prefixes while applying real-time regex secrets redaction. |
| `-parallel` | `int` | `NumCPU` | Maximum number of concurrent tasks executed across the worker pool (e.g., `-parallel 8`). |
| `-dry-run` | `bool` | `false` | Previews topological execution batches and dependency levels without executing any commands. |
| `-no-fail-fast` | `bool` | `false` | Continues executing independent sibling tasks even if a task fails. Ideal for CI test collection. |
| `-json` | `bool` | `false` | Emits structured JSON summary of execution times, cache hit ratios, and task statuses to stdout. |
| `-provenance` | `bool` | `false` | Generates SLSA v1.0 verifiable JSON-LD build provenance attestation file. |
| `-config` | `string` | `tasks.json` | Path to a custom tasks configuration file. |
| `-quiet` | `bool` | `false` | Suppresses the startup banner and ASCII art in terminal output. |
| `--` | `args` | N/A | Pass-through delimiter. All subsequent arguments are forwarded directly to the task's command line. |

---

## 6. `tasks.json` Configuration Specification

ZEPHYR is configured via a standard JSON file (`tasks.json`) in the project root:

```json
{
  "default": "build",
  "tasks": {
    "lint": {
      "command": "go vet ./...",
      "inputs": ["**/*.go"],
      "outputs": [],
      "timeout": "1m"
    },
    "build": {
      "command": "go build -o bin/taskrunner-demo main.go",
      "inputs": ["main.go", "go.mod"],
      "outputs": ["bin/taskrunner-demo", "bin/taskrunner-demo.exe"],
      "depends_on": ["lint"],
      "env_vars": ["GOOS", "GOARCH", "APP_ENV"],
      "timeout": "2m"
    },
    "repro-build": {
      "command": "go build -trimpath -ldflags=-buildid= -o bin/taskrunner-demo.exe main.go",
      "inputs": ["main.go", "go.mod"],
      "outputs": ["bin/taskrunner-demo.exe"],
      "depends_on": ["lint"],
      "timeout": "2m"
    },
    "test:unit": {
      "command": "go test -v -run TestDAG ./...",
      "inputs": ["**/*.go"],
      "outputs": [],
      "depends_on": ["lint"],
      "timeout": "1m",
      "retries": 1
    },
    "test:integration": {
      "command": "go test -v -run TestPipeline ./...",
      "inputs": ["**/*.go", "tasks.json"],
      "outputs": [],
      "depends_on": ["build", "test:unit"],
      "timeout": "2m"
    },
    "package": {
      "command": "go version",
      "inputs": ["bin/taskrunner-demo*", "README.md"],
      "outputs": [],
      "depends_on": ["test:integration"],
      "env_vars": ["APP_ENV"]
    }
  }
}
```

### Schema Property Reference

| Property | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `command` | `string` | **Yes** | Shell command executed inside the task workspace. |
| `inputs` | `string[]` | No | Glob patterns for files determining the task's cache fingerprint (e.g. `["src/**/*.ts"]`). |
| `outputs` | `string[]` | No | File or directory paths saved to CAS and restored on cache hits. Supports globs like `dist/*`. |
| `depends_on` | `string[]` | No | List of upstream tasks that must complete successfully before this task runs. |
| `env_vars` | `string[]` | No | Environment variable names whose values are incorporated into the task cache key. |
| `timeout` | `string` | No | Execution timeout (e.g., `"30s"`, `"2m"`). Kills deep child process trees on expiry. |
| `retries` | `integer` | No | Number of retry attempts on non-zero exit codes. |
| `weight` | `integer` | No | Relative concurrency cost (defaults to 1). |
| `cwd` | `string` | No | Subdirectory path where the command executes (validated to remain inside workspace). |

### Polyglot Monorepo Example

ZEPHYR seamlessly orchestrates multi-language projects within a single unified DAG:

```json
{
  "default": "all",
  "tasks": {
    "go:lint": {
      "command": "golangci-lint run ./...",
      "inputs": ["services/**/*.go", "go.mod"],
      "cwd": "services"
    },
    "go:build": {
      "command": "go build -o ../bin/api main.go",
      "inputs": ["services/**/*.go", "go.mod"],
      "outputs": ["bin/api"],
      "depends_on": ["go:lint"],
      "cwd": "services"
    },
    "web:typecheck": {
      "command": "pnpm tsc --noEmit",
      "inputs": ["apps/web/src/**/*.{ts,tsx}", "apps/web/tsconfig.json"],
      "cwd": "apps/web"
    },
    "web:build": {
      "command": "pnpm vite build",
      "inputs": ["apps/web/src/**/*.{ts,tsx,css,html}", "apps/web/package.json"],
      "outputs": ["apps/web/dist/*"],
      "depends_on": ["web:typecheck"],
      "cwd": "apps/web"
    },
    "all": {
      "command": "echo Pipeline Complete",
      "depends_on": ["go:build", "web:build"]
    }
  }
}
```

---

## 7. Architecture & Engineering Innovations

### 7.1. Two-Tier CAS with 256-Bucket Sharding
Traditional caching systems store blobs in a single directory. In large codebases (>100,000 files), this creates severe B+ tree search degradation on NTFS and HTree depth expansion on Linux ext4.

ZEPHYR shards blobs across 256 hexadecimal buckets (`.taskcache/cas/objects/ab/cdef...`), keeping directory structures within kernel dentry cache limits and maintaining **$O(1)$ lookups (~15–35 µs)** even at 1,000,000 objects.

### 7.2. Zero-Copy Hardlink Restoration vs. In-Memory Compression
* **The Empirical Experiment:** We tested compressing cached blobs using `compress/gzip`. While storage footprint dropped by 18.4%, decompression added **14.2ms of CPU latency per task**.
* **The Solution:** Local blobs are stored uncompressed and linked into workspaces using OS-level `os.Link` hardlinks. Output restoration completes in **0.08ms at 0% CPU cost**, sharing physical disk inodes without duplicating storage.

### 7.3. Dynamic Critical Path Method (CPM) Priority Scheduling
Naive topological execution processes parallel tasks in arbitrary order. If a long-running compile task is delayed behind trivial linting jobs, total build duration suffers.

ZEPHYR logs task completion times in `.taskcache/stats.json` and calculates the Critical Path length:

$$\text{CP}(T) = \text{Duration}(T) + \max_{D \in \text{Dependents}(T)} \text{CP}(D)$$

Tasks that dictate the overall pipeline completion time are scheduled first, reducing multicore build times by **20%–35%**.

### 7.4. Speculative Early Cutoff (Hash-Invariant Pruning)
In traditional build systems, updating an upstream task (e.g. changing a comment in a shared library) causes all downstream dependents to rebuild.

ZEPHYR computes the Merkle output digest of generated artifacts:
1. If the upstream task runs and produces byte-identical outputs, ZEPHYR flags `[EARLY CUTOFF]`.
2. Downstream tasks receive the invariant output digest token.
3. Downstream tasks bypass execution and remain **100% CACHED**.

### 7.5. Process Tree Termination on Windows and Unix
Terminating long-running tasks or timed-out processes on Windows frequently leaves grandchild processes running in the background, holding open stdout/stderr pipes and deadlocking build orchestrators.

ZEPHYR solves this using:
* `cmd.Cancel`: Invokes `taskkill /F /T /PID` on Windows or `syscall.Kill(-pgid, syscall.SIGKILL)` on Unix.
* `cmd.WaitDelay`: Configured with a 2-second grace period to forcefully reap hung pipes.

---

## 8. Empirical Benchmarks & Performance

Measured using the empirical multi-trial benchmark runner (`zephyr bench`) testing actual Go compilations across $N=5$ trials:

*Environment: AMD Ryzen 5 7600X (12 Cores) | 32 GB DDR5 RAM | Windows 11 AMD64 | NVMe SSD | Go 1.27.0*

| Benchmark Condition | Median Duration | Min Duration | Max Duration | P95 Duration | Measured Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Clean Execution (Cold)** | **1,378.7 ms** | 1,365.1 ms | 2,095.4 ms | 2,095.4 ms | Baseline |
| **Cache Hit Retrieval (Warm)**| **1.24 ms** | 1.12 ms | 4.42 ms | 4.42 ms | **1,108.9× Faster** |
| **1-File Diff (Incremental)**| **2,061.9 ms** | 2,010.5 ms | 2,234.3 ms | 2,234.3 ms | Targeted Rebuild |
| **Capsule Replay (Zero-Cache)**| **28.7 ms** | 24.1 ms | 31.5 ms | 31.5 ms | Instant Extract |

* **Avoided Execution:** 60.0% of tasks avoided on single-file modifications.
* **Reproducibility Digest:** Byte-identical across clean runs (`72a5dff4...`).

---

## 9. Security & Supply-Chain Governance

1. **Live Stream Secrets Redaction:** Stdout and stderr streams pass through a real-time regex redaction filter, masking GitHub tokens (`ghp_`), Stripe secret keys (`sk_live_`), and Authorization headers (`Bearer ...`) with `***REDACTED***`.
2. **Tar Slip Jail Protection:** All archive unpackers (`capsule replay`, tar extractors) validate relative paths against target destinations, refusing directory traversal attempts (`../`).
3. **Hardened Remote HTTP Cache:** Default localhost binding (`127.0.0.1`), mandatory bearer token authentication, and strict payload size limits (`http.MaxBytesReader`).
4. **SLSA v1.0 Build Provenance:** `zephyr run --provenance` generates verifiable JSON-LD build provenance documents capturing input hashes, builder identity, and environment configurations.

---

## 10. Automated Testing & Verification Suites

### 10.1. System Verification Suite (`verify.ps1`)
Runs 13 end-to-end checks validating zero dependencies, byte-identical builds, doctor checks, cache speedups, and capsule workflows:

```powershell
powershell -ExecutionPolicy Bypass -File .\verify.ps1
```

### 10.2. Master Adversarial & Chaos Test Harness (`test/run_all_tests.ps1`)
Executes 14 stress and edge-case suites testing deep DAG recursion, race conditions, zombie processes, unicode file handling, and monorepo circularity:

```powershell
powershell -ExecutionPolicy Bypass -File .\test\run_all_tests.ps1
```

```text
==============================================================================
ZEPHYR 2.0 MASTER TEST SUITE EXECUTION SUMMARY (ALL 14 SUITES)
==============================================================================
Suite                                               Status Duration ExitCode
-----                                               ------ -------- --------
Suite 01: DAG and Topology Stress                   PASS   31.26s          0
Suite 02: Security, Tar Slip and Anti-Tamper        PASS   5.14s           0
Suite 03: Polyglot Project Adoption                 PASS   2.95s           0
Suite 04: CAS Storage and Caching Engine            PASS   7.29s           0
Suite 05: Concurrency, Timeouts and Retries         PASS   23.33s          0
Suite 06: AI Agent Interface and MCP Protocol       PASS   2.46s           0
Suite 07: Reproducibility Auditor and Diffing       PASS   3.95s           0
Suite 08: Cache Concurrency Race Conditions         PASS   1.52s           0
Suite 09: Hardlink In-Place Mutation Trap           PASS   0.45s           0
Suite 10: Deep Process Tree Termination             PASS   1.62s           0
Suite 11: Real-Time Stream Redaction Masking        PASS   0.42s           0
Suite 12: Output Globbing in 'dist/' Dir            PASS   0.41s           0
Suite 13: Filesystem Edge Cases (Zero-Byte/Unicode) PASS   0.48s           0
Suite 14: Monorepo Cross-Package Circularity        PASS   0.43s           0

Total Suites: 14 | Passed: 14 | Failed: 0 (100% Pass)
```

### 10.3. Go Standard Package Tests
```bash
go test -v ./...
```

---

## 11. Competitive Comparison Matrix

| Feature / Architecture | Turborepo | Nx | Just / Make | **ZEPHYR 2.0 ⚡** |
| :--- | :--- | :--- | :--- | :--- |
| **External Dependencies** | Rust / Node.js binaries | Heavy NPM tree | Varies | **ZERO (100% Go Stdlib)** |
| **Binary Footprint** | ~35 MB + Node runtime | ~120 MB + NPM | ~2 MB | **~4.5 MB Static Executable** |
| **Reproducibility Auditor**| None | None | None | **`zephyr verify` (Byte-offset diff)** |
| **Cryptographic Capsules** | Basic Tar | Basic Tar | None | **Ed25519 Signed (`.zcap`)** |
| **Zero-Config Adoption** | Manual configuration | Plugin based | None | **`zephyr adopt` (Go/Node/Rust/Py/Make)** |
| **AI Agent Interface** | None | None | None | **`zephyr agent` & `zephyr mcp`** |
| **Workspace Diagnostics** | None | `nx report` | None | **`zephyr doctor` (5-category audit)** |
| **Cache Storage Engine** | Tar in `.turbo` | Tar archives | None | **Two-Tier CAS ($0\text{ms}$ Hardlinks)** |
| **CAS Scalability** | Flat directory | Flat directory | None | **256-Bucket Sharded CAS (`cas/ab/cd...`)** |
| **Scheduling Optimization**| Naive topological | Topological | Topological | **Dynamic Critical Path Method (CPM)** |
| **Downstream Invalidation**| Full blind rebuild | Full blind rebuild| Full rebuild | **Speculative Early Cutoff** |
| **Cache-Miss Diagnostics** | Basic | Verbose diff | None | **Intelligent `--why` Explanations** |
| **Stream Secrets Masking** | None | None | None | **Live Stream Redaction (`***REDACTED***`)** |
| **Supply-Chain Attestation**| Custom JSON | None | None | **SLSA v1.0 Provenance Attestations** |

---

## 12. Zero-Dependency Proof

Verify that ZEPHYR imports zero third-party dependencies:

```bash
go list -m all
```
*Output:*
```text
taskrunner
```

Inspect `go.mod`:
```text
module taskrunner

go 1.22
```
*(Empty `require` block. 100% Go Standard Library).*

For an exhaustive audit of all standard library architecture substitutions, see the [Standard Library Substitution Log](docs/STDLIB.md).

### Go Standard Library Substitution Map

| Feature Area | Typical Third-Party Library | ZEPHYR 2.0 Go Standard Library Engine |
| :--- | :--- | :--- |
| **CLI Argument Parsing** | `spf13/cobra`, `urfave/cli` | `flag.FlagSet` + custom sub-command dispatch |
| **Cryptographic Signatures**| `crypto/rsa`, `golang.org/x/crypto`| `crypto/ed25519` + `crypto/sha256` |
| **Archive Containers** | `archiver`, `mholt/archiver` | `archive/tar` + `compress/gzip` + jail guards |
| **File Globbing** | `bmatcuk/doublestar` | `path/filepath.WalkDir` + `path.Match` |
| **AI Agent Protocol** | `modelcontextprotocol/sdk` | `encoding/json` + `bufio.Scanner` (JSON-RPC 2.0 over stdio) |
| **Remote HTTP Cache** | `gin-gonic/gin`, `gorilla/mux`| `net/http` + `http.MaxBytesReader` |
| **Concurrent Workflows** | `sourcegraph/conc` | Native `sync.WaitGroup` + buffered worker channels |

---

## 13. Author & License

* **Author:** Naman Swami
* **Project:** ZEPHYR 2.0 (Local-First Build Intelligence Engine)
* **Repository:** [https://github.com/naman-swami/ZEPHYR-2.0.git](https://github.com/naman-swami/ZEPHYR-2.0.git)
* **License:** MIT License — see [LICENSE](LICENSE) for details.
