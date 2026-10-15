# ZEPHYR 2.0 ⚡ — Verified, Local-First Build Intelligence

**A zero-third-party-dependency Go tool for executing task graphs, explaining cache decisions, auditing reproducibility, and verifying build artifacts.**

---

## 💡 Inspiration

Modern software projects rely on repeated operations such as compiling code, running tests, checking quality, and packaging releases. As projects grow and AI coding assistants make code changes more frequently, developers need to understand not only how to execute these operations, but also why a task ran, whether its outputs are reproducible, and whether a previously generated artifact can be trusted.

I started ZEPHYR to explore a simple question:

**Can we make software builds more transparent, reproducible, and efficient without requiring a large additional toolchain?**

Existing build systems already solve important parts of this problem. Tools like Nx and Turborepo offer affected-task analysis and remote caching, while systems like Bazel provide strict hermetic sandboxing. ZEPHYR takes a different implementation approach: a lightweight, local-first engine written in Go using strictly the Go standard library, bringing task orchestration, content-based caching, execution diagnostics, reproducibility auditing, and signed artifact capsules into a single self-contained binary.

ZEPHYR 2.0 extends that foundation toward a broader goal: helping developers and AI coding agents understand, inspect, and verify the computations behind software changes.

---

## ⚙️ What It Does

ZEPHYR is an incremental build and task orchestration engine built around a directed acyclic graph (DAG) of tasks and their dependencies.

Its core capabilities include:

**1. Incremental Execution and Cache Diagnostics**
ZEPHYR calculates task fingerprints from declared inputs, environment variables, and configuration, reuses eligible cached results from a two-tier content-addressable store (CAS), and executes dependent tasks in topological order. Parallel execution and affected-task analysis help avoid unnecessary work. The `--why` diagnostic explains cache misses so developers can investigate why a task needs to run again instead of treating the build system as a black box.

**2. Reproducibility Auditing — `zephyr verify`**
The verification command runs a task repeatedly under controlled conditions in isolated workspace copies and compares exit codes, output streams, and declared output artifacts. When binary outputs differ, it diagnoses the divergence down to the exact byte offset (e.g. `Offset 0x280 (640)`). The goal is to make reproducibility measurable rather than assumed, with results explicitly scoped to the inputs, environment, and runs tested.

**3. Portable Build Capsules — `zephyr capsule`**
Build Capsules package execution metadata, SHA-256 digests, and build outputs into portable `.zcap` archives. Developers can inspect a capsule, verify its recorded hashes, and restore its packaged outputs without relying on the original local task cache. Where publisher signatures are enabled, asymmetric Ed25519 signature verification provides an additional cryptographic guarantee that a capsule originated from a trusted key, while lexical path containment prevents directory-traversal vulnerabilities during extraction.

**4. Project Discovery — `zephyr adopt`**
Project discovery reduces the manual effort needed to configure a build system. Adapters inspect common project manifests and generate inspectable task configurations for Go (`go.mod`), Node.js (`package.json`), Rust (`Cargo.toml`), Python (`pyproject.toml`), and Makefiles, with safe overwrite protection.

**5. Structured AI-Agent Integration — `zephyr agent` & `zephyr mcp`**
AI coding agents need more than the ability to edit files: they need to query the tasks affected by an edit, inspect execution results, and avoid triggering blind full rebuilds. ZEPHYR exposes structured JSON interfaces and a native Model Context Protocol (MCP 2024-11-05) stdio server (`zephyr_graph`, `zephyr_affected`, `zephyr_why`, `zephyr_verify`, `zephyr_run`). Command execution is permission-gated so an agent cannot execute processes through the interface without explicit human authorization (`--allow-exec`).

**6. Developer Environment & System Integrations**
The ZEPHYR 2.0 prototype also incorporates a built-in Language Server Protocol (`zephyr lsp`) server for in-editor cycle detection and autocomplete, peer-to-peer local network cache synchronization (`zephyr p2p`) using UDP discovery and verified CAS sync, and kernel-level task resource profiling (`zephyr profile`) generating standalone interactive SVG Flame Graphs and Chrome/Perfetto execution traces.

---

## 🧭 Command Architecture & Discovery Matrix: Engineering Problems That Motivated Each Capability

Every command in ZEPHYR was engineered to address a specific practical friction point encountered in modern software development and build maintenance:

| Command & Flags | Primary Capabilities | Engineering Problem That Led to Its Creation | Implementation & Guarantees |
| :--- | :--- | :--- | :--- |
| **`zephyr run [targets]`**<br>`--parallel`, `--why`, `--dry-run`, `--stream`, `--json`, `--fail-fast` | Concurrently executes DAG tasks, prioritizes bottleneck paths using historical durations (CPM), and streams output with real-time secret redaction. | **Opaque Rebuilds & Log Exposure:** Developers frequently encounter unexpected rebuilds without understanding what triggered invalidation. `--why` prints exact cryptographic diffs (`file modified: path [hash A -> hash B]`), while `--stream` masks common secret tokens (`ghp_`, `sk_live_`, `Bearer`) in live logs. | `os/exec`, `sync`, `regexp`, `crypto/sha256`, `time` |
| **`zephyr verify <task>`**<br>`--runs=N` (default: 2) | Executes a task across repeated clean workspace clones without cache reuse and compares output streams and artifacts byte-for-byte. | **Diagnosing Non-Deterministic Output:** Build outputs frequently differ across clean environments due to embedded timestamps (`time.Now()`), compiler build IDs, or toolchain paths. `verify` pinpoints the first byte divergence offset (`0x280`) to isolate non-deterministic compiler flags. | `os.Link`, `io`, `crypto/sha256`, `bytes.Compare` |
| **`zephyr capsule`**<br>`keygen`, `create`, `verify`, `replay`, `inspect` | Packages task outputs, input digests, and metadata into portable archives with optional Ed25519 asymmetric signatures. | **Tamper Detection & Publisher Authentication:** Sharing compiled artifacts across machines requires distinguishing content integrity from publisher identity. `.zcap` capsules combine SHA-256 artifact hashing with Ed25519 publisher signatures and lexical path containment jails (`filepath.Clean`). | `crypto/ed25519`, `archive/tar`, `compress/gzip`, `filepath` |
| **`zephyr adopt`**<br>`--force` | Scans workspace manifests and generates an initial `tasks.json` configuration for Go, Node.js, Rust, Python, or Make projects. | **Manual Migration Overhead:** Onboarding existing polyglot codebases to a new task runner often requires writing extensive repetitive configuration. `adopt` infers dependencies and outputs from established project files in a single pass. | `encoding/json`, `bufio`, `os`, `path/filepath` |
| **`zephyr agent`**<br>`graph`, `affected`, `verify`, `run --allow-exec` | Emits structured JSON envelopes for DAG inspection, change blast-radius analysis, and permission-gated execution. | **AI Agent Execution Safety:** Autonomous AI coding agents can trigger uncontrolled full rebuilds or destructive terminal commands. `agent` provides read-only graph queries and requires explicit human consent (`--allow-exec`) before spawning commands. | `encoding/json`, `os/exec`, `context` |
| **`zephyr mcp`**<br>*(stdio JSON-RPC 2.0)* | Implements the Model Context Protocol over stdio for direct integration into AI editors (Claude Code, Cursor). | **Standardizing Build Context for AI Assistants:** AI assistants lack native semantic tools to inspect build graphs or verify cache hits. `mcp` exposes five dedicated tools without requiring external network ports or background services. | `bufio.Scanner`, `encoding/json`, `os.Stdin/Stdout` |
| **`zephyr lsp`**<br>`--check`, `--config` | Provides Language Server Protocol (LSP 3.17) diagnostics, DAG cycle detection, duration hovers, and schema autocompletion. | **Catching Configuration Cycles During Editing:** Syntax errors and circular dependency deadlocks in build configurations are typically only caught at runtime. `lsp` surfaces errors directly in the editor as code is typed. | `bufio.Reader`, `encoding/json`, `net/textproto` |
| **`zephyr p2p`**<br>`status`, `serve`, `discover`, `sync` | Discovers local subnet peers via UDP broadcast and synchronizes verified CAS objects over HTTP with HMAC authentication. | **Low-Overhead LAN Cache Sharing:** Small teams working on local networks often lack dedicated cloud storage buckets for remote caching. `p2p` enables direct peer-to-peer cache synchronization with mandatory SHA-256 verification of received blobs. | `net.UDPConn`, `crypto/hmac`, `crypto/subtle`, `net/http` |
| **`zephyr profile <task>`**<br>`--svg`, `--trace` | Measures process user/system CPU time and peak memory, generating standalone SVG Flame Graphs and Chrome Trace JSON. | **Low-Overhead Pipeline Profiling:** Diagnosing build bottlenecks usually requires separate profiling tools that add overhead. `profile` uses standard kernel process statistics to visualize task execution and resource allocation. | `os.ProcessState`, `strings.Builder`, `encoding/json` |
| **`zephyr doctor`** | Inspects local development prerequisites, toolchain availability, write permissions, and OS shell configuration. | **Pre-Flight Environment Validation:** Build failures often stem from missing system utilities or misconfigured paths. `doctor` validates the execution environment before running tasks. | `os`, `os/exec`, `runtime` |
| **`zephyr clean`** | Safely removes local Action Cache manifests and CAS objects while preserving historical runtime statistics. | **Cache Invalidation & Storage Management:** Accumulated build artifacts consume disk space over time. `clean` purges cached outputs without discarding duration records used by CPM scheduling. | `os.RemoveAll`, `path/filepath` |
| **`zephyr affected`**<br>`--files=path1,path2` | Computes the minimal subgraph of tasks impacted by a specific list of modified file paths. | **Avoiding Redundant Task Invocations:** Editing documentation or a single sub-package should not trigger an entire pipeline. `affected` isolates only downstream dependent targets. | Topological Traversal, Adjacency Lists |
| **`zephyr graph` / `list`** | Renders an ASCII visualization of the task dependency graph or outputs a tabular summary of configured tasks. | **Understanding Complex Task Relationships:** Deeply nested task graphs can be difficult to interpret from raw JSON. `graph` provides a visual representation of execution order and dependency paths. | Recursive DFS tree traversal, `fmt` |
| **`zephyr bench`** | Executes repeated controlled trials across clean and warm cache conditions to measure execution timings. | **Verifiable Performance Measurement:** Performance claims in developer tooling require reproducible evaluation. `bench` provides a self-contained multi-trial test harness with median, min, and max timings. | Controlled trial loops, `time.Now`, statistical aggregations |
| **`zephyr server`**<br>`--port`, `--token`, `--dir` | Serves an HTTP remote cache endpoint with constant-time token validation and request body limits. | **Lightweight Shared Caching:** Teams needing a centralized cache often have to deploy complex external storage infrastructure. `server` provides a zero-dependency remote cache server runnable anywhere. | `net/http`, `crypto/subtle`, `http.MaxBytesReader` |
| **`zephyr watch [targets]`** | Continuous file watch mode with 350ms topological debounce and reverse-DAG affected-task rebuild. | **Editor Save Storms & Blind Rebuilds:** Naive file watchers fire duplicate builds on editor atomic writes (e.g. vim temp swap/rename) and rebuild the whole project. `watch` debounces events and rebuilds only the affected reverse-DAG subgraph. | `time.Ticker`, `filepath.WalkDir`, `time.Sleep` |
| **`zephyr stress`** | Automated Chaos Gauntlet testing 6 extreme architectural stress and resilience scenarios. | **Unverified Resilience Claims:** Build systems rarely prove resilience against bit rot, kill-storm interrupts, or path fuzzing. `stress` tests corruption quarantine, crash-atomic writes, 1,000-node DAG cycles, clock skew, and Tar Slip traversal in automated reproducible suites. | `crypto/sha256`, `math/rand`, `sync`, `os.Chtimes` |

---

## 🛠️ How We Built It

ZEPHYR is implemented in Go and relies exclusively on the Go standard library for its own execution, avoiding third-party Go module dependencies (an empty `require ()` block in `go.mod` across 11 internal packages and 28 standard library subsystems).

Its architecture combines several established computer science techniques:

- **Dependency graphs:** Topological ordering (Kahn's algorithm) and three-color depth-first search (DFS) cycle detection establish valid execution sequences.
- **Content-based caching & Self-Healing CAS:** Streaming SHA-256 hashing computes input fingerprints; a two-tier 256-bucket sharded CAS manages deduplication with crash-atomic `fsync` writes, verify-on-read corruption detection, and automatic quarantine.
- **Concurrent execution:** Goroutine pools and channel synchronization schedule independent tasks concurrently while respecting dependency constraints.
- **Cryptographic verification:** SHA-256 digests provide content integrity, while asymmetric Ed25519 signatures authenticate capsule publishers.
- **Structured interfaces:** Stdout JSON-RPC 2.0 servers provide integration points for the Model Context Protocol (MCP) and Language Server Protocol (LSP).
- **Portable packaging & Safety Jails:** Build Capsules package selected build outputs and manifest metadata using standard `archive/tar` and `compress/gzip` with lexical path jails defending against Tar Slip attacks.
- **Chaos & Resilience Engine:** Self-contained stress test engine (`pkg/stress`) rigorously exercising atomic crash safety, memory corruption quarantine, and 1,000-node graph stress.

The zero-third-party-dependency design applies strictly to ZEPHYR's Go implementation. Individual projects orchestrated by ZEPHYR still use whichever compilers, runtimes, package managers, and toolchains their own tasks require.

---

## 🧪 Challenges We Ran Into

**Reproducibility is more complicated than comparing successful exit codes.**  
Two commands can both succeed while producing different artifacts. We therefore compare output streams and declared output binaries byte-for-byte, reporting divergence offsets rather than treating a zero exit code as proof of determinism.

**Cache integrity and publisher trust are different problems.**  
A cryptographic hash detects discrepancies against a recorded digest, but an asymmetric digital signature is required to verify publisher identity. Keeping those guarantees separate is critical when artifacts move across untrusted networks or shared caches.

**Automatic project discovery must be conservative.**  
A detected script does not automatically reveal every dependency, environment variable, or output it uses. Generated task graphs are designed as safe starting points that developers can inspect, refine, and commit.

**Performance measurements need context.**  
A warm cache lookup performs an entirely different amount of work than a full compiler invocation. We built a benchmark harness to report controlled execution timings across explicit conditions, with the aim of making measurements transparent, repeatable, and verifiable.

---

## 📊 What We Measured

In a controlled five-trial experiment on Windows 11 / AMD64, using an AMD Ryzen 5 7600X (12 Cores, 4.7 GHz) and 32 GB of DDR5 RAM with Go 1.27.0, the observed median timings were:

| Workload Condition | Median Duration | Min Duration | Max Duration | Characterization |
| :--- | :--- | :--- | :--- | :--- |
| **Clean Execution (Cold)** | **1,378.7 ms** | 1,365.1 ms | 2,095.4 ms | Baseline full compilation of task pipeline |
| **Warm-Cache Retrieval** | **1.24 ms** | 1.12 ms | 4.42 ms | Metadata check & hardlink CAS artifact restoration |
| **One-File-Difference Run** | **2,061.9 ms** | 2,010.5 ms | 2,234.3 ms | Selective recompilation of affected target subgraph |
| **Build Capsule Replay** | **28.7 ms** | 24.1 ms | 31.5 ms | Zero-cache unarchive, verification & output extraction |

*Methodological Scope Disclosure:* These are results from a specific benchmark environment (`zephyr bench`), not universal performance claims across arbitrary projects. Cold execution and warm-cache retrieval perform fundamentally different workloads; the incremental rebuild duration depends directly on which tasks within the DAG are affected by a given code modification. In this test pipeline, three out of five tasks were skipped during the tested one-file modification, demonstrating task-skipping efficiency for that specific graph topology.

---

## 🏆 What We Accomplished

- Built a single-binary Go task execution engine with dependency-aware scheduling and incremental caching using 100% Go Standard Library.
- Developed an automated reproducibility auditor (`zephyr verify`) that tests output consistency across clean workspace clones.
- Implemented portable Build Capsules (`.zcap`) with cryptographic hash verification, Ed25519 publisher signatures, and Tar Slip extraction protection.
- Added zero-config project discovery (`zephyr adopt`) and structured AI agent interfaces (`zephyr agent` & `zephyr mcp`).
- Implemented protocol extensions including an in-editor Language Server (`zephyr lsp`), local subnet cache sharing (`zephyr p2p`), and kernel-level resource profiling (`zephyr profile`).
- Created the Chaos Gauntlet (`zephyr stress`): 6 automated architectural stress scenarios testing bit-rot corruption quarantine, crash-atomic fsync, 1,000-node DAG cycles, clock skew invariance, and Tar Slip traversal fuzzing.
- Established a comprehensive verification harness: 17/17 verification checks passing in `verify.ps1`, 14/14 adversarial stress test suites passing in `test/run_all_tests.ps1`, and 11/11 Go packages passing unit tests.

---

## 📚 What We Learned

Build speed is only one part of reliable software infrastructure. Developers and AI assistants also need transparent explanations for execution decisions, evidence regarding output determinism, and clear guarantees about the authenticity of restored artifacts.

We also learned that security and reproducibility claims must be precisely defined. Hash verification, publisher authentication, repeated-build consistency, and operating-system isolation are distinct engineering properties; one should never be presented as a substitute for another.

---

## 🚀 What's Next

Our next priorities are to validate ZEPHYR on larger and more varied repositories, strengthen adversarial security testing, publish repeatable benchmark procedures, and improve integration with developer environments and AI coding agents.

The long-term direction is a build intelligence layer that helps developers understand the impact of code changes and gives automated tools a more reliable way to plan, execute, inspect, and verify software tasks.

**ZEPHYR 2.0 is an exploration of a simple principle: builds should be not only efficient, but also explainable and verifiable.**
