# ZEPHYR 2.0 ⚡: Verified, Local-First Build Intelligence

**A Go-based build and task orchestration tool that explains cache decisions, audits reproducibility, and packages verifiable build artifacts with no third-party Go module dependencies.**

## 💡 Inspiration

Software projects repeatedly compile code, run tests, check quality, and package releases. As projects grow and AI coding assistants make code changes more frequently, developers need more than a command that runs the build. They need to understand **why a task ran, what a change affects, whether repeated runs produce the same outputs, and whether restored artifacts can be trusted**.

I started ZEPHYR to explore a simple question:

**Can software builds become more transparent, reproducible, and efficient without requiring another large toolchain?**

Existing build systems already solve important parts of this problem. Tools like Nx and Turborepo offer affected-task analysis and remote caching, while systems like Bazel provide strict hermetic sandboxing. ZEPHYR explores a different implementation approach: a lightweight, local-first engine written in Go using strictly the Go standard library, combining dependency-aware task execution, content-based caching, execution diagnostics, reproducibility checks, and portable build capsules into a single self-contained binary.

The long-term vision is to give developers and AI coding agents a clearer, more verifiable understanding of the computations behind software changes.

## ⚙️ What ZEPHYR Does

ZEPHYR models tasks and their dependencies as a directed acyclic graph (DAG). It executes independent tasks concurrently, reuses eligible cached results, and identifies tasks affected by selected file changes.

### 1. Incremental execution with explanations

ZEPHYR fingerprints declared task inputs and configuration, stores reusable artifacts in a content-addressable store (CAS), and schedules tasks according to their dependencies. The `--why` diagnostic explains cache misses and shows why a task must run again, rather than leaving developers to guess why a build was invalidated.

### 2. Reproducibility auditing: zephyr verify

The verifier runs a task repeatedly in separate workspace copies without reusing the task cache, then compares exit codes, output streams, and declared output artifacts byte for byte. If outputs differ, ZEPHYR reports the divergence, including the first differing byte offset for binary comparisons (for example, `Offset 0x280 (640)`).

This makes reproducibility something that can be tested. Results apply to the particular task, inputs, environment, and runs tested; they are not a guarantee that every environmental dependency has been discovered.

### 3. Portable Build Capsules: zephyr capsule

Build Capsules package selected task metadata, SHA-256 digests, and output artifacts into portable `.zcap` archives. The commands support creating, inspecting, verifying, and restoring capsule contents without depending on the original local task cache.

Where Ed25519 signing is enabled and the signer's public key is trusted, signature verification establishes publisher authenticity. Hash verification and signature verification serve distinct purposes: hashes check content against recorded digests, while signatures associate the manifest with a signing key. Capsule extraction also enforces lexical path containment to prevent directory traversal vulnerabilities.

### 4. Project discovery: zephyr adopt

The `adopt` command inspects common project manifests and proposes a starting task configuration for Go, Node.js, Rust, Python, and Makefile projects. The generated configuration remains inspectable, because discovered scripts do not automatically reveal every input, dependency, environment variable, or output a task may use.

### 5. Structured AI agent interface: zephyr agent and zephyr mcp

AI coding agents can change code quickly, but they still need to understand a project's task graph and the consequences of an edit. ZEPHYR exposes structured information for graph inspection, affected-task analysis, cache explanations, and verification results.

The MCP stdio interface lets compatible AI development environments query build context without requiring a separate network service. Agent-triggered process execution is gated behind an explicit `--allow-exec` opt-in to ensure human oversight.

### 6. Continuous watching, chaos testing, and developer tools

The ZEPHYR 2.0 prototype also incorporates specialized developer tooling:
- `zephyr watch`: continuous reverse-DAG watcher with 350ms topological debounce, rebuilding only invalidated tasks and their downstream dependents.
- `zephyr stress`: multi-scenario Chaos Gauntlet testing corruption quarantine, crash-atomic writes, 1,000-node DAG cycles, clock skew invariance, and Tar Slip traversal fuzzing.
- `zephyr lsp`: Language Server Protocol server for in-editor cycle detection, hover telemetry, and autocomplete.
- `zephyr p2p`: peer-to-peer local network cache synchronization over UDP discovery and verified CAS sync.
- `zephyr profile`: kernel-level task resource profiling generating interactive SVG Flame Graphs and Chrome/Perfetto execution traces.
- `zephyr doctor` and `zephyr server`: workspace pre-flight validation and an authenticated HTTP remote cache server.

## 🛠️ How We Built It

ZEPHYR is implemented in Go and relies exclusively on the Go standard library for its own execution, avoiding third-party Go module dependencies (an empty `require ()` block in `go.mod` across 11 internal packages and 28 standard library subsystems). Projects orchestrated by ZEPHYR still use whichever compilers, runtimes, package managers, and toolchains their own tasks require.

Its architecture combines several established computer science techniques:

- **Dependency graph analysis:** topological ordering (Kahn's algorithm) and three-color depth-first search (DFS) cycle detection establish valid task execution order.
- **Content-based caching and self-healing storage:** streaming SHA-256 hashing computes input fingerprints; a two-tier 256-bucket sharded CAS manages deduplication with crash-atomic fsync writes, verify-on-read corruption detection, and automatic quarantine into `.taskcache/quarantine/`.
- **Concurrent scheduling:** goroutines and channel synchronization coordinate independent tasks while respecting dependency constraints and configured resource limits.
- **Reproducibility checks:** repeated executions compare exit status, output streams, and declared artifact bytes.
- **Cryptographic verification:** SHA-256 digests provide content integrity, while optional Ed25519 signatures authenticate a capsule's signer when the public key is trusted.
- **Structured protocols:** JSON-RPC 2.0 stdio interfaces connect the task engine to IDE tooling (LSP) and compatible AI agent clients (MCP).
- **Portable packaging and safety jails:** standard library archive tools package selected outputs and metadata, protected by lexical path containment jails defending against Tar Slip attacks.
- **Chaos and resilience engine:** automated stress testing (`pkg/stress`) exercising bit-rot corruption quarantine, crash safety, and 1,000-node graph stress.

The zero third-party dependency design is a property of ZEPHYR's Go implementation, not a claim that the projects it builds need no external tools.

## 🧭 Key Commands and the Problems They Address

| Command | Purpose | Problem addressed | Implementation and guarantees |
|---|---|---|---|
| `zephyr run` | Executes selected tasks and their dependencies; supports parallel execution, CPM scheduling, and diagnostics. | Unnecessary rebuilds, slow sequential pipelines, and unclear cache decisions. | `os/exec`, `sync`, `regexp`, `crypto/sha256` |
| `zephyr watch` | Continuous file watcher with 350ms topological debounce and reverse-DAG rebuilds. | Editor save storms firing duplicate builds and naive full rebuilds on single-file edits. | `time.Ticker`, `filepath.WalkDir`, `time.Sleep` |
| `zephyr affected` | Identifies and executes tasks affected by selected file changes or git diffs. | Running unrelated tasks after a small edit. | Reverse-DAG traversal, adjacency lists |
| `zephyr verify` | Repeats execution across clean sandboxes and compares output bytes. | Assuming a successful exit code means a reproducible build. | Clean directory sandboxing, byte-offset diffing |
| `zephyr capsule` | Creates, inspects, verifies, and restores portable `.zcap` archives. | Moving build artifacts without a portable record of their expected contents or provenance. | `crypto/ed25519`, `archive/tar`, path jail |
| `zephyr adopt` | Generates a starting task configuration from project manifests. | Manual setup when trying a task orchestrator in an existing repository. | Polyglot scanners for Go, Node, Rust, Python, Make |
| `zephyr agent` / `zephyr mcp` | Exposes structured build context to automation and compatible AI tools. | AI tools making changes without understanding task dependencies or triggering runaway commands. | JSON-RPC 2.0 over stdio, permission-gated execution |
| `zephyr stress` | Automated Chaos Gauntlet testing 6 extreme architectural stress and resilience scenarios. | Unverified resilience claims regarding bit rot, sudden process kills, or path fuzzing. | 6 automated fault and fuzzing scenarios in `pkg/stress` |
| `zephyr lsp` / `zephyr profile` | Provides editor diagnostics and kernel-level task profiling outputs. | Discovering configuration problems or execution bottlenecks too late. | LSP 3.17 stdio server, SVG Flame Graph generator |
| `zephyr doctor` / `zephyr server` / `zephyr p2p` | Checks the environment and supports shared-cache workflows. | Environment drift and friction when sharing reusable build artifacts across teams. | Pre-flight audits, HTTP remote cache, P2P LAN sync |

## 🧪 Challenges We Ran Into

**A successful command is not necessarily a reproducible build.** Two runs can both exit successfully while producing different outputs. The verifier therefore compares actual outputs under the conditions tested instead of using exit status alone as evidence.

**Cache integrity is not the same as publisher trust.** A SHA-256 digest can detect content that differs from a recorded digest, but it cannot by itself prove who created the artifact. Signed capsules add a separate trust mechanism, provided the verification key is trusted and the signature is checked.

**Project discovery must be conservative.** A manifest can reveal useful commands, but not every hidden dependency or side effect. Automatically generated task graphs should be treated as inspectable starting points, not infallible descriptions of a build.

**Performance numbers need context.** A warm-cache retrieval does less work than a complete compilation. We built a benchmark harness to record repeatable timings under specific conditions and to avoid presenting unlike workloads as direct performance equivalents.

## 📊 What We Measured

The reported five-trial experiment ran on Windows 11 / AMD64, using an AMD Ryzen 5 7600X (12 Cores, 4.7 GHz), 32 GB of DDR5 RAM, and Go 1.27.0. The measured medians were:

| Workload | Median | Minimum | Maximum | Characterization |
|---|---:|---:|---:|---|
| Clean execution (cold) | 1,378.7 ms | 1,365.1 ms | 2,095.4 ms | Baseline full compilation of task pipeline |
| Warm-cache retrieval | 1.24 ms | 1.12 ms | 4.42 ms | Metadata check and hardlink CAS artifact restoration |
| One-file-difference run | 2,061.9 ms | 2,010.5 ms | 2,234.3 ms | Selective recompilation of affected target subgraph |
| Build Capsule replay | 28.7 ms | 24.1 ms | 31.5 ms | Zero-cache unarchive, verification and output extraction |

In the tested five-task graph, three tasks were skipped after the one-file modification, demonstrating task-skipping efficiency for that specific graph topology. These figures describe this benchmark setup, not expected performance for all repositories. The warm cache and cold execution cases do different amounts of work, so their ratio should not be treated as a general build speedup. The measured one-file-difference run was slower than the cold baseline in this experiment; it does not demonstrate an incremental speedup and is a result to investigate in further benchmarking.

The reproducibility experiment also reported byte-identical outputs across the tested clean workspace copies. This is evidence for that experiment's conditions, not proof of universal determinism. No energy or carbon savings are claimed because dedicated power measurements were not part of this experiment.

## 🏆 What We Accomplished

- Built a single-binary Go task execution engine with dependency-aware scheduling and incremental caching.
- Added an automated reproducibility auditor (`zephyr verify`) that compares repeated task outputs under controlled conditions.
- Implemented portable Build Capsules (`.zcap`) with cryptographic hash verification, Ed25519 publisher signatures, and Tar Slip extraction protection.
- Added project discovery (`zephyr adopt`) and structured interfaces for automation and AI-agent integration (`zephyr agent` and `zephyr mcp`).
- Extended the prototype with in-editor Language Server Protocol support (`zephyr lsp`), kernel-level resource profiling (`zephyr profile`), and peer-to-peer cache sharing (`zephyr p2p`).
- Created the Chaos Gauntlet (`zephyr stress`): 6 automated architectural stress scenarios testing bit-rot corruption quarantine, crash-atomic fsync, 1,000-node DAG cycles, clock skew invariance, and Tar Slip traversal fuzzing.
- Expanded automated tests across the core engine and all 11 internal packages: `taskrunner`, `adopt`, `agent`, `bench`, `capsule`, `lsp`, `mcp`, `p2p`, `profile`, `stress`, and `verify`.
- Completed 17/17 automated end-to-end verification checks in `verify.ps1` and 14/14 adversarial stress test suites in `test/run_all_tests.ps1`.

All results are reproducible from the exact source revision submitted with the project using 100% Go Standard Library with zero third-party dependencies.

## 📚 What We Learned

Build speed is only one part of reliable software infrastructure. Developers also need explanations for execution decisions, evidence about output consistency, and clear information about the trustworthiness of restored artifacts.

We learned to distinguish several properties that are easy to conflate: content integrity, publisher authentication, repeated build consistency, and operating system isolation. Each needs its own implementation and tests; none should be presented as a substitute for another.

## 🚀 What's Next

The next priorities are to evaluate ZEPHYR on larger and more varied repositories, publish the benchmark harness and exact run instructions, expand adversarial security testing, and improve integrations with developer environments and AI coding agents.

The long-term direction is a build intelligence layer that helps developers understand the impact of code changes and gives automation a more reliable way to plan, execute, inspect, and verify software tasks.

**ZEPHYR 2.0 explores a simple principle: builds should be not only efficient, but also explainable and verifiable.**
