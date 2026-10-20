# ZEPHYR 2.0 ⚡: Verified, Local First Build Intelligence

**A Go based build and task orchestration tool that explains cache decisions, audits reproducibility, signs build provenance, and packages verifiable artifacts with no third party Go module dependencies.**

## 💡 Inspiration

Every save triggers the same loop: compile, test, lint, package. As projects grow, and as AI coding assistants start making those changes for us, a command that merely runs the build is no longer enough. Developers need to know **why a task ran, what a change affects, whether repeated runs produce the same outputs, and whether a restored artifact can be trusted, all the way down to who signed it.**

I started ZEPHYR to answer one question:

**Can software builds become transparent, reproducible, and efficient without bolting on another large toolchain?**

Existing build systems already solve parts of this. Nx and Turborepo offer affected task analysis and remote caching. Bazel provides strict hermetic sandboxing, at the cost of a heavy toolchain and its own build language. ZEPHYR takes a different path: a lightweight, local first engine written entirely with the Go standard library, combining dependency aware execution, content based caching, execution diagnostics, reproducibility checks, signed provenance, and portable build capsules in one self contained binary.

The goal is a build tool that developers, and the AI agents working beside them, can actually interrogate: every decision explained, every artifact checkable, every signature verifiable.

## ⚙️ What ZEPHYR Does

ZEPHYR models tasks and their dependencies as a directed acyclic graph (DAG). It executes independent tasks concurrently, reuses eligible cached results, and identifies tasks affected by selected file changes.

### 1. Incremental execution with explanations

ZEPHYR fingerprints declared task inputs and configuration, stores reusable artifacts in a content addressable store (CAS), and schedules tasks according to their dependencies. The `--why` diagnostic explains cache misses and shows why a task must run again, instead of leaving developers to guess why a build was invalidated.

A task purity gate protects all of this: tasks that produce side effects, such as deploying or writing outside their workspace, are marked impure and can never seed a cached result or feed the speculative early cutoff. A side effect can therefore never poison the cache.

### 2. Reproducibility auditing: zephyr verify

The verifier runs a task repeatedly in separate workspace copies without reusing the task cache, then compares exit codes, output streams, and declared artifacts byte for byte. If outputs differ, ZEPHYR reports the divergence down to the first differing byte offset for binary comparisons (for example, `Offset 0x280 (640)`). Reproducibility becomes something you can test instead of something you assume.

### 3. Portable Build Capsules with signed provenance: zephyr capsule

Build Capsules package selected task metadata, SHA-256 digests, and output artifacts into portable `.zcap` archives that can be created, inspected, verified, and restored without the original local cache.

New in 2.0, every capsule can carry signed supply chain provenance: a SLSA Provenance v1 predicate carried in an In-toto Statement v1 payload, serialized by a deterministic canonical JSON encoder and wrapped in a DSSE envelope (Dead Simple Signing Envelope) using PAE, the Pre Authentication Encoding. Signing uses an Ed25519 key ring with key rotation and revocation, so a compromised key can be retired without reissuing every capsule.

Hashes and signatures answer different questions: hashes check content against recorded digests, while signatures tie the manifest to a signing key. Extraction is hardened as well: the extractor rejects symlinks, hardlinks, device nodes, and name collision tricks, and the path jail blocks null bytes, percent encoded traversals like `%2e%2e%2f`, and Windows reserved device names such as `CON` and `NUL`.

### 4. Project discovery: zephyr adopt

The `adopt` command inspects common project manifests and proposes a starting task configuration for Go, Node.js, Rust, Python, and Makefile projects. The result stays inspectable, because a discovered script does not reveal every hidden input, environment variable, or side effect. Adopted configs are starting points, not infallible descriptions of a build.

### 5. Structured AI agent interface: zephyr agent and zephyr mcp

AI coding agents can change code quickly, but they still need to understand a project's task graph and the consequences of an edit. ZEPHYR exposes structured information for graph inspection, affected task analysis, cache explanations, and verification results over an MCP stdio interface, with no separate network service. Agent triggered process execution is gated behind an explicit `--allow-exec` opt in, so a human stays in the loop.

### 6. Crash safe cache management: zephyr prune

Deleting from a cache is where naive strategies quietly destroy live state. The obvious approach, evict oldest entries until under a size limit, can delete a blob that a live manifest still references, or one that a build started seconds ago is about to request. The pruning engine runs a two phase mark and sweep with referential integrity: any blob referenced by an active manifest is never deleted. Eviction is crash atomic, keeps 10% headroom so the cache does not thrash, and grants in flight builds a 15 minute grace period.

### 7. Hardened worker pool and live telemetry

The scheduler defends the machine it runs on. A token bucket caps the subprocess spawn rate so the kernel process table cannot be exhausted, a weighted memory governor reads Linux cgroup v2, v1, and meminfo before launching new work, and an aging aware priority queue keeps long jobs from starving. A live Prometheus `/metrics` endpoint plus OTLP telemetry expose queue depth, memory pressure, and task timings while builds run.

### 8. Continuous watching, chaos testing, and developer tools

- `zephyr watch`: continuous reverse DAG watcher with 350ms topological debounce, rebuilding only invalidated tasks and their downstream dependents.
- `zephyr stress`: the Chaos Gauntlet, 6 automated scenarios including corruption quarantine, crash atomic writes, 1,000 node DAG cycles, clock skew invariance, and Tar Slip traversal fuzzing. All 6 pass.
- `zephyr lsp`: Language Server Protocol server for in editor cycle detection, hover telemetry, and autocomplete.
- `zephyr p2p`: peer to peer local network cache sync over UDP discovery and verified CAS sync, with bounded quarantine forensics and replay protection on beacons.
- `zephyr profile`: kernel level task resource profiling generating interactive SVG Flame Graphs and Chrome/Perfetto execution traces.
- `zephyr doctor` and `zephyr server`: workspace pre flight validation and an authenticated HTTP remote cache server.

### 9. Self healing code repairs and build diagnostics: zephyr heal and zephyr diagnose

Modern AI coding agents and human developers inevitably encounter syntax errors, indentation mistakes, and build failures. ZEPHYR 2.0 introduces a local first dual tier developer intelligence engine:

- `zephyr heal <file>`: Analyzes compiler error diagnostics, invokes our fine tuned sub 1B multi task model (`Qwen2.5-Coder-0.5B`) to propose a minimal unified diff repair, and gates the candidate patch through a deterministic AST compiler gate (`go/parser` for Go, `ast.parse` for Python, and delimiter validation for JavaScript). The patch is committed only if it passes AST parsing with zero errors, preventing malformed edits. Every repair is recorded in `.taskcache/provenance/heal_audit.json` with SHA-256 pre patch and post patch hashes and signed provenance.
- `zephyr diagnose`: Accepts raw build logs or compiler stderr (via arguments or piped stdin) and performs instant root cause analysis on cyclic task graphs, concurrency data races, missing toolchains, or proxy network blocks, returning actionable next steps.

### 10. Workspace organization and AI token optimizer: zephyr organize

Bulky, flat, or disorganized codebases waste 75% to 88% of LLM context window tokens and cause AI agents to hallucinate file paths or enter circular build loops. `zephyr organize` automatically analyzes repository layout, groups peripheral assets into standard directories (`docs/`, `assets/`, `scripts/`, `configs/`), and uses Abstract Syntax Tree import analysis to cluster code files into functional domains.

Five mandatory safety gates protect the codebase:
- Clean Git check prevents modifying uncommitted workspaces.
- Dry run previews display relocation tables before any disk mutation.
- Tarjan's strongly connected components (SCC) cycle gate prevents introducing circular package dependencies.
- A two phase transaction rollback journal (`.taskcache/reorg/latest.json`) logs every move with SHA-256 pre and post state digests.
- Automatic compiler verification (`go vet` or `zephyr doctor`) verifies build validity and triggers an instant sub 50 millisecond rollback if any error occurs.

### 11. Windows File Explorer address bar terminal: zephyr shell and zephyr install-shell

Developers often navigate workspaces using Windows File Explorer. Running `zephyr install-shell` registers ZEPHYR in Windows `App Paths` and adds an Explorer context menu item (`Open ZEPHYR Terminal Here`). Typing `zephyr` directly into the File Explorer address bar launches an interactive terminal session in that specific folder.

Using native Win32 `kernel32.dll` console process inspection (`GetConsoleProcessList`), ZEPHYR detects when File Explorer spawned a dedicated console window and automatically launches an interactive REPL rather than abruptly flashing and closing. Furthermore, each terminal session binds strictly to its local workspace directory, allowing developers to run multiple concurrent ZEPHYR shells across different microservices simultaneously without state or lock contention.

### 12. AI credential isolation and shadow vault: zephyr vault

Autonomous AI coding agents (such as Claude Code, Cursor, GitHub Copilot, Devin, or OpenHands) inspect configuration files to write client adapters or tests, frequently leaking private API keys, database credentials, or auth tokens into remote prompt logs. `zephyr vault` establishes a local credential isolation boundary:

- `zephyr vault audit`: Scans workspace configuration for exposed credentials, high entropy secrets, and private keys.
- `zephyr vault cloak <file>`: Generates a syntactically valid cloaked configuration where sensitive credentials are substituted with deterministic synthetic tokens (`ZEPHYR_VAULT_SECRET_<hash>`). The AI agent receives complete schema structure without seeing cleartext secrets.
- `zephyr vault merge <target> <proposed>`: When an AI agent modifies configuration or adds new environment variables, the three way reconciliation engine parses both ASTs, imports newly added variables, and guarantees that user authentic credentials are preserved intact without being overwritten or leaked.

## 🛠️ How We Built It

ZEPHYR is implemented in Go and relies exclusively on the Go standard library for its own execution: the `require ()` block in `go.mod` is empty. The codebase spans 17 internal packages, now including `pkg/heal` for self healing AST verification gates, `pkg/organize` for automated workspace layout, `pkg/shell` for Explorer terminal integration, `pkg/vault` for AI credential isolation, `pkg/prune` for cache lifecycle, and `pkg/sched` for the worker pool and telemetry. Projects orchestrated by ZEPHYR still use whatever compilers, runtimes, and package managers their own tasks require.

Its architecture combines established systems techniques:

- **Dependency graph analysis:** topological ordering (Kahn's algorithm) and three color depth first search cycle detection establish valid execution order.
- **Content based caching and self healing storage:** streaming SHA-256 hashing computes input fingerprints; a two tier, 256 bucket sharded CAS manages deduplication with crash atomic fsync writes (same shard temp files, directory fsync for ext4 delayed allocation, and retry with exponential backoff for Windows file locking), verify on read corruption detection, and automatic quarantine into `.taskcache/quarantine/`.
- **Concurrent scheduling:** goroutines and channels coordinate independent tasks under dependency and resource limits.
- **Reproducibility checks:** repeated executions compare exit status, output streams, and artifact bytes.
- **Cryptographic verification:** SHA-256 digests provide content integrity; Ed25519 signatures over In-toto Statement v1 payloads carrying SLSA Provenance v1, wrapped in DSSE envelopes, authenticate the publisher.
- **Structured protocols:** JSON RPC 2.0 over stdio connects the engine to IDE tooling (LSP) and AI agent clients (MCP).
- **Safety jails:** lexical path containment defends extraction against Tar Slip attacks, null bytes, percent encoded traversals, and Windows reserved device names.
- **Chaos and resilience engine:** `pkg/stress` exercises bit rot quarantine, crash safety, and 1,000 node graph stress.

## 🧭 Key Commands and the Problems They Address

| Command | Purpose | Problem addressed | Implementation and guarantees |
|---|---|---|---|
| `zephyr run` | Executes selected tasks and their dependencies; supports parallel execution, CPM scheduling, and diagnostics. | Unnecessary rebuilds, slow sequential pipelines, and unclear cache decisions. | `os/exec`, `sync`, `regexp`, `crypto/sha256` |
| `zephyr watch` | Continuous file watcher with 350ms topological debounce and reverse DAG rebuilds. | Editor save storms firing duplicate builds and naive full rebuilds on single file edits. | `time.Ticker`, `filepath.WalkDir`, `time.Sleep` |
| `zephyr affected` | Identifies and executes tasks affected by selected file changes or git diffs. | Running unrelated tasks after a small edit. | Reverse DAG traversal, adjacency lists |
| `zephyr verify` | Repeats execution across clean sandboxes and compares output bytes. | Assuming a successful exit code means a reproducible build. | Clean directory sandboxing, byte offset diffing |
| `zephyr capsule` | Creates, inspects, verifies, and restores portable `.zcap` archives with signed provenance. | Moving build artifacts without a portable record of their expected contents or provenance. | `crypto/ed25519`, `archive/tar`, DSSE provenance, path jail |
| `zephyr prune` | Crash safe cache eviction with referential integrity. | Unbounded caches, or prunes that delete blobs still in use. | Two phase mark and sweep in `pkg/prune` |
| `zephyr adopt` | Generates a starting task configuration from project manifests. | Manual setup when trying a task orchestrator in an existing repository. | Polyglot scanners for Go, Node, Rust, Python, Make |
| `zephyr agent` / `zephyr mcp` | Exposes structured build context to automation and compatible AI tools. | AI tools making changes without understanding task dependencies or triggering runaway commands. | JSON RPC 2.0 over stdio, permission gated execution |
| `zephyr stress` | Automated Chaos Gauntlet running 6 fault injection and fuzzing scenarios. | Unverified resilience claims regarding bit rot, sudden process kills, or path fuzzing. | 6 automated fault and fuzzing scenarios in `pkg/stress`, all passing |
| `zephyr lsp` / `zephyr profile` | Provides editor diagnostics and kernel level task profiling outputs. | Discovering configuration problems or execution bottlenecks too late. | LSP 3.17 stdio server, SVG Flame Graph generator |
| `zephyr heal` | Self healing code repairs gated by deterministic AST compiler validation. | Syntax errors, indentation mistakes, and malformed edits breaking builds. | Sub 1B model, `go/parser`, `ast.parse`, signed SHA-256 provenance audit |
| `zephyr diagnose` | Root cause analysis on build, linker, concurrency, and DAG failures. | Deciphering cryptic compiler stack traces and multi process build failures. | AI diagnostics engine, piped stdin, automated remediation advice |
| `zephyr organize` | Automated workspace layout & AI token optimizer with 5 safety gates. | Bulky repository file sprawl, LLM context window exhaustion, and build thrashing. | Two tier lexical and semantic engine, Tarjan's SCC cycle check, atomic rollback journal |
| `zephyr shell` / `zephyr install-shell` | Windows File Explorer address bar terminal and multi instance launcher. | Context switching to dedicated terminal windows and terminal flash and close behavior. | Win32 console process guard (`GetConsoleProcessList`), Explorer `App Paths` registration |
| `zephyr vault` | AI credential isolation, synthetic cloaking, and 3 way secret reconciler. | Private API keys and database credentials leaking into AI agent prompts. | High entropy secret scanner, deterministic synthetic tokens, AST preserving merger |
| `zephyr doctor` / `zephyr server` / `zephyr p2p` | Checks the environment and supports shared cache workflows. | Environment drift and friction when sharing reusable build artifacts across teams. | Pre flight audits, HTTP remote cache, P2P LAN sync |

## 🧪 Challenges We Ran Into

**A successful command is not necessarily a reproducible build.** Two runs can both exit cleanly while producing different outputs, so the verifier compares actual outputs instead of trusting exit status.

**Cache integrity is not the same as publisher trust.** A SHA-256 digest can detect that content differs from a recorded digest, but it cannot prove who created the artifact. Signed provenance adds that separate mechanism, and it only means something if the verification key is trusted.

**Deleting from a cache is genuinely dangerous.** The obvious prune strategy can delete a blob that a live manifest still references, or one that a build started seconds ago is about to request. Two phase mark and sweep with referential integrity, headroom, and grace periods is what it takes to make deletion safe.

**Operating systems disagree about durability.** A rename that is atomic on one file system can fail outright across devices, ext4 delayed allocation means a directory entry may not survive a crash without an explicit directory fsync, and Windows throws sharing violations on rename. The CAS now writes same shard temp files, fsyncs directories, and retries with exponential backoff.

**Project discovery must be conservative.** A manifest reveals useful commands but not every hidden dependency or side effect, so adopted graphs are inspectable starting points.

**Performance numbers need context.** A warm cache retrieval does less work than a full compilation, so we built a benchmark harness to record repeatable timings under stated conditions and avoid presenting unlike workloads as equivalents.

## 📊 What We Measured

The experiment ran five trials on Windows 11 / AMD64 (AMD Ryzen 5 7600X, 6 cores and 12 threads, 32 GB DDR5, Go 1.27.0). Measured medians:

| Workload | Median | Minimum | Maximum | Characterization |
|---|---:|---:|---:|---|
| Clean execution (cold) | 1,378.7 ms | 1,365.1 ms | 2,095.4 ms | Baseline full compilation of task pipeline |
| Warm cache retrieval | 1.24 ms | 1.12 ms | 4.42 ms | Metadata check and hardlink CAS artifact restoration |
| One file difference run | 2,061.9 ms | 2,010.5 ms | 2,234.3 ms | Selective recompilation of affected target subgraph |
| Build Capsule replay | 28.7 ms | 24.1 ms | 31.5 ms | Unarchive from a capsule with no local cache, digest verification, and output extraction |

In the tested five task graph, three tasks were skipped after a one file edit. One honest surprise: the one file difference run came out slower than the cold baseline in this setup, so it does not demonstrate an incremental speedup, and it is first on the list for further benchmarking. All figures describe this benchmark setup, not expected performance for other repositories, and the warm cache and cold cases do different amounts of work. The reproducibility experiment reported byte identical outputs across the tested workspace copies. No energy or carbon savings are claimed, because power was not measured.

## 🏆 What We Accomplished

- Built a single binary Go task execution engine with dependency aware scheduling, incremental caching, and a purity gate that keeps side effects out of the cache.
- Shipped the reproducibility auditor (`zephyr verify`) comparing repeated outputs under controlled conditions.
- Shipped portable Build Capsules (`.zcap`) with SHA-256 verification, Ed25519 publisher signatures, SLSA Provenance v1 inside In-toto Statement v1 payloads wrapped in DSSE envelopes, a key ring with rotation and revocation, and Tar Slip hardened extraction.
- Shipped crash safe cache pruning (`pkg/prune`) with referential integrity, anti thrashing headroom, and in flight grace periods.
- Shipped the hardened worker pool (`pkg/sched`) with spawn rate limiting, a cgroup aware memory governor, starvation free aging, and a live Prometheus `/metrics` endpoint with OTLP telemetry.
- Added project discovery (`zephyr adopt`) and structured AI agent integration (`zephyr agent`, `zephyr mcp`).
- Extended the prototype with in editor LSP support, kernel level profiling, and peer to peer cache sharing.
- Shipped automated workspace layout and AI token optimizer (`zephyr organize`) with two tier classification, Tarjan's SCC cycle prevention, and atomic rollback.
- Shipped Windows File Explorer address bar terminal launcher and multi instance isolation (`zephyr shell`, `zephyr install-shell`) with Win32 console process protection.
- Shipped AI credential isolation and shadow vault (`zephyr vault`) with synthetic cloaking and 3 way AST secret reconciliation.
- Expanded automated tests across all 17 internal packages: `taskrunner`, `adopt`, `agent`, `bench`, `capsule`, `heal`, `lsp`, `mcp`, `organize`, `p2p`, `profile`, `prune`, `sched`, `shell`, `stress`, `vault`, and `verify`.
- Completed 17/17 automated end to end verification checks in `verify.ps1`, 14/14 adversarial suites in `test/run_all_tests.ps1`, and 6/6 Chaos Gauntlet scenarios.

All results are reproducible from the exact source revision submitted with the project, using 100% Go standard library and zero third party dependencies.

## 📚 What We Learned

The fastest cache in the world is useless if you cannot tell why a task ran, whether two runs matched, or who signed the artifact you just restored. Speed is one property; explanation, consistency, and trust are separate properties, each needing its own implementation and its own tests.

We also learned that the unglamorous work is where builds actually break: a prune that deletes a live blob, a rename that is not durable on ext4, a cache entry poisoned by a side effect. None of these show up in a demo, and all of them show up in production.

## 🚀 What's Next

Evaluate ZEPHYR on larger and more varied repositories, publish the benchmark harness and exact run instructions, expand adversarial security testing, and deepen integrations with developer environments and AI coding agents.

The long term direction is a build intelligence layer that helps developers understand the impact of code changes and gives automation a reliable way to plan, execute, inspect, and verify software tasks.

ZEPHYR 2.0 rests on one bet: a build tool earns trust when it can explain itself and prove what it produced.
