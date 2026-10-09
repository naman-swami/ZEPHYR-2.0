# ZEPHYR 2.0 — The 4-Minute Winning Demo Script
### EurekaDev 2026 Hackathon Video Presentation Guide
**Author & Presenter:** Naman Swami  
**Target Video Duration:** 3:45 – 4:00  
**Video Resolution:** 1080p / 60fps (Recommended terminal font: JetBrains Mono / Cascadia Code, size 16-18pt)

---

## Quick Reference / Command Sequence Cheat Sheet

```powershell
# 1. Clean Slate & Discovery (0:00 - 0:45)
.\zephyr.exe doctor
.\zephyr.exe adopt

# 2. Cold Execution vs 0ms Replay (0:45 - 1:30)
.\zephyr.exe clean
.\zephyr.exe run
.\zephyr.exe run

# 3. Targeted Change & --why Diagnostic (1:30 - 2:10)
# (Make 1-line edit or touch file)
.\zephyr.exe run --why

# 4. Reproducibility Auditor (2:10 - 2:50)
.\zephyr.exe verify build
.\zephyr.exe verify repro-build

# 5. Ed25519 Build Capsules (2:50 - 3:30)
.\zephyr.exe capsule keygen
.\zephyr.exe capsule create --key zephyr.key repro-build
.\zephyr.exe capsule inspect .taskcache/capsules/repro-build.zcap
.\zephyr.exe capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
.\zephyr.exe capsule replay --out-dir demo_restore .taskcache/capsules/repro-build.zcap

# 6. AI Agent Integration & Zero-Dep Proof (3:30 - 4:00)
.\zephyr.exe agent graph
.\zephyr.exe agent affected --files=main.go
go list -m all
```

---

## Detailed Scene-by-Scene Script & Narration

### Scene 1: The Hook & Project Discovery (0:00 – 0:45)
**Screen Setup:** Clean terminal, high-contrast dark theme (cyberpunk/dracula), terminal width ~100 columns.

**Narration:**
> "Hi, I'm Naman Swami, and this is **ZEPHYR 2.0** — a local-first Build Intelligence Engine engineered entirely with the Go Standard Library: **zero external dependencies**, zero `node_modules`, and zero package bloat.
>
> In the age of AI coding assistants, code is generated faster than ever. But our build tools are fundamentally blind: they rely on fragile timestamps, leak phantom non-determinism, and force developers to waste hours waiting for CI to rebuild code that never changed.
>
> ZEPHYR transforms builds from reactive script execution into **verified computation intelligence**."

**Action on Screen:**
```powershell
.\zephyr.exe doctor
```
*Visual Hook:* Terminal displays the Cybernetic ASCII HUD, auditing workspace health, DAG validity, and hardware cores.

**Action on Screen:**
```powershell
.\zephyr.exe adopt
```
**Narration:**
> "Notice how ZEPHYR adopts an existing project instantly with zero configuration. It inspects Go, Node, Rust, Python, and Makefiles, extracts build rules, and builds a dependency Directed Acyclic Graph automatically."

---

### Scene 2: Cold Execution vs. 0ms Hardlink Replay (0:45 – 1:30)
**Action on Screen:**
```powershell
.\zephyr.exe clean
.\zephyr.exe run
```
*Visual Hook:* Pipeline executes `lint` and `build` in real time with Braille runes and live duration timers (~4-6 seconds).

**Narration:**
> "Here is a cold build. ZEPHYR executes the pipeline, hashes every source file and environment variable with SHA-256, and indexes output artifacts into a two-tier Content-Addressable Storage (CAS)."

**Action on Screen:**
```powershell
.\zephyr.exe run
```
*Visual Hook:* Immediate instantaneous return! 
```
Summary : 2 cached, 0 executed, 0 failed
Duration: 14ms (Compute saved: 6.7s / 99.8%)
```

**Narration:**
> "Now, watch what happens when we rerun the build. **14 milliseconds.** 99.8% of compute saved. ZEPHYR uses OS-level hardlinks to restore output artifacts in microseconds with zero byte copying and zero memory overhead."

---

### Scene 3: The `--why` Invalidation Detective (1:30 – 2:10)
**Action on Screen:**
Add a small comment or touch `main.go`, then run:
```powershell
.\zephyr.exe run --why
```
*Visual Hook:* Terminal displays:
```
[build] [MISS] Modified inputs: main.go (hash changed: e67ec5 -> 88b12f)
```

**Narration:**
> "Every developer knows the frustration of cache blindness: *'Why did my task rerun?'*
>
> With `--why`, ZEPHYR diffs the exact input fingerprint against the cached manifest, immediately highlighting whether a source file was touched, an environment variable drifted, or a compiler flag changed."

---

### Scene 4: The Reproducibility Auditor (`zephyr verify`) (2:10 – 2:50)
**Narration:**
> "Now for our most important engineering contribution: **zephyr verify**. Two builds producing the same output today does not prove a build is reproducible tomorrow.
>
> ZEPHYR spawns clean, isolated workspace clones, executes tasks without cache, and compares output artifacts byte-for-byte."

**Action on Screen (Show failure detection first):**
```powershell
.\zephyr.exe verify build
```
*Visual Hook:* Terminal outputs:
```
══════════════════════════════════════════════════════════════════════════════
  ZEPHYR REPRODUCIBILITY AUDIT REPORT
  VERDICT: [FAIL] NON-DETERMINISTIC BEHAVIOR DETECTED
    ✗ bin\taskrunner-demo [HASH DIVERGENCE]
      Run 1 SHA-256: 701064c01a...
      Run 2 SHA-256: 2fa0692f20...
      First Byte Divergence: Offset 0x280 (640)
  Audit Findings: Embedded compile timestamps or build IDs detected.
══════════════════════════════════════════════════════════════════════════════
```

**Narration:**
> "Standard `go build` embeds build timestamps and temporary path IDs. ZEPHYR immediately catches the divergence, pinpointing the exact byte offset where the binaries differ (offset `0x280`)."

**Action on Screen (Show reproducible pass):**
```powershell
.\zephyr.exe verify repro-build
```
*Visual Hook:*
```
  VERDICT: [PASS] 100% REPRODUCIBLE UNDER TESTED CONDITIONS
    ✓ Exit Codes Match: 0
    ✓ Stdout Stream: Byte-identical across runs
    ✓ bin\taskrunner-demo.exe: SHA-256 MATCH (92284c9911...)
```

**Narration:**
> "When reproducible build flags are supplied, ZEPHYR certifies 100% determinism across independent workspaces."

---

### Scene 5: Ed25519 Tamper-Resistant Build Capsules (2:50 – 3:30)
**Narration:**
> "Once an artifact is verified, how do you distribute it safely without trusting an untrusted network? We introduce **Build Capsules** (`.zcap`)."

**Action on Screen:**
```powershell
.\zephyr.exe capsule create --key zephyr.key repro-build
.\zephyr.exe capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
```
*Visual Hook:*
```
[CAPSULE VERIFY] PASS: All 1 bundled artifacts match recorded SHA-256 checksums.
  • Signature Authenticity: Ed25519 Authenticated (Signer: aa7811592d928dda...)
```

**Narration:**
> "Capsules bundle the execution manifest, output artifacts, and cryptographic provenance into a single archive signed with an asymmetric **Ed25519** key. During replay, ZEPHYR validates publisher authenticity, checks artifact hashes, and enforces strict **Tar Slip path traversal jails** to prevent malicious archive extraction."

**Action on Screen:**
```powershell
.\zephyr.exe capsule replay --out-dir demo_restore .taskcache/capsules/repro-build.zcap
```

---

### Scene 6: AI Agent Semantic Interface & Zero-Dependency Proof (3:30 – 4:00)
**Action on Screen:**
```powershell
.\zephyr.exe agent graph
.\zephyr.exe agent affected --files=main.go
```
*Visual Hook:* Clean, streaming JSON output for AI coding agents.

**Narration:**
> "Finally, ZEPHYR is built for the AI era. Through `zephyr agent`, tools like GitHub Copilot, Replit Agent, or Claude can query the dependency graph and compute affected blast radiuses via structured JSON. And execution is strictly gated behind explicit authorization (`--allow-exec`)."

**Action on Screen:**
```powershell
go list -m all
```
*Visual Hook:* Terminal outputs strictly:
```
taskrunner
```

**Narration:**
> "And the crowning achievement? An empty `require` block. Zero third-party dependencies. 100% Go Standard Library.
>
> That is **ZEPHYR 2.0**: The Local-First Build Intelligence Engine. Thank you!"

---

## Presenter Tips for Max Scores

1. **Energy & Clarity:** Speak with confidence. Do not rush; let the terminal animations breathe.
2. **Highlight the Contrast:** The moment `zephyr verify` flags `Offset 0x280` in red and then `repro-build` passes in bright green is the technical peak of the video. Linger on that screen for 3 seconds.
3. **EurekaDev Tiebreaker Strategy:** Reiterate **Impact & Relevance** — *"AI proposes code changes; ZEPHYR deterministically predicts, verifies, and replays the computation."*
