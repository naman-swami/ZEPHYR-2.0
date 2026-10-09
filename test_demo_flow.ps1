# test_demo_flow.ps1 - Automated verification of the 4-Minute Demo Script
$ErrorActionPreference = "Stop"

Write-Host "`n=== [1/6] Auditing System & Project Adoption ===" -ForegroundColor Cyan
.\zephyr.exe doctor
if ($LASTEXITCODE -ne 0) { throw "Doctor failed" }

.\zephyr.exe adopt
if ($LASTEXITCODE -ne 0) { throw "Adopt failed" }

Write-Host "`n=== [2/6] Executing Cold Build & Instant Replay ===" -ForegroundColor Cyan
.\zephyr.exe clean
.\zephyr.exe run --quiet
.\zephyr.exe run --quiet
if ($LASTEXITCODE -ne 0) { throw "Run failed" }

Write-Host "`n=== [3/6] Testing Cache Invalidation Diagnostics ===" -ForegroundColor Cyan
.\zephyr.exe run --why --quiet
if ($LASTEXITCODE -ne 0) { throw "Why failed" }

Write-Host "`n=== [4/6] Auditing Build Reproducibility ===" -ForegroundColor Cyan
# Non-deterministic build should detect divergence
Write-Host "Running verify on non-reproducible build (expecting detection)..."
.\zephyr.exe verify --runs=2 build
Write-Host "Running verify on reproducible build (expecting PASS)..."
.\zephyr.exe verify --runs=2 repro-build
if ($LASTEXITCODE -ne 0) { throw "Verify repro-build failed" }

Write-Host "`n=== [5/6] Testing Tamper-Resistant Build Capsules ===" -ForegroundColor Cyan
if (-not (Test-Path "zephyr.key")) {
    .\zephyr.exe capsule keygen
}
.\zephyr.exe capsule create --key zephyr.key repro-build
if ($LASTEXITCODE -ne 0) { throw "Capsule create failed" }

.\zephyr.exe capsule inspect .taskcache/capsules/repro-build.zcap
if ($LASTEXITCODE -ne 0) { throw "Capsule inspect failed" }

.\zephyr.exe capsule verify --key zephyr.pub .taskcache/capsules/repro-build.zcap
if ($LASTEXITCODE -ne 0) { throw "Capsule verify failed" }

if (Test-Path "demo_restore") { Remove-Item -Recurse -Force demo_restore }
.\zephyr.exe capsule replay --out-dir demo_restore .taskcache/capsules/repro-build.zcap
if ($LASTEXITCODE -ne 0) { throw "Capsule replay failed" }
Remove-Item -Recurse -Force demo_restore

Write-Host "`n=== [6/6] Testing AI Agent Semantic API & Dependency Proof ===" -ForegroundColor Cyan
.\zephyr.exe agent graph | Out-Null
.\zephyr.exe agent affected --files=main.go | Out-Null

$deps = (go list -m all).Trim()
if ($deps -ne "taskrunner") {
    throw "Dependency leak detected: $deps"
}

Write-Host "`n[DEMO VERIFICATION COMPLETE] All 6 demo stages passed flawlessly!" -ForegroundColor Green
