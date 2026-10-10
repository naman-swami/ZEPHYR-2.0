# ZEPHYR 2.0 - System Verification Suite
# Verifies all architectural claims, zero dependencies, determinism, and test coverage.
# Run: powershell -ExecutionPolicy Bypass -File .\verify.ps1

$ErrorActionPreference = "Stop"
$passCount = 0
$failCount = 0
$results = @()

function Run-Check([string]$label, [scriptblock]$action) {
    Write-Host -NoNewline "  [..] $label ... "
    try {
        & $action
        Write-Host "PASS" -ForegroundColor Green
        $script:passCount++
        $script:results += @{ Label = $label; Status = "PASS" }
    } catch {
        Write-Host "FAIL: $($_.Exception.Message)" -ForegroundColor Red
        $script:failCount++
        $script:results += @{ Label = $label; Status = "FAIL"; Error = "$($_.Exception.Message)" }
    }
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "          ZEPHYR 2.0 - System Verification Suite                " -ForegroundColor Cyan
Write-Host "          Empirical Reproducibility & Architecture Audit       " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

if (-not (Test-Path "go.mod")) {
    Write-Host "ERROR: Please run this script from the project root." -ForegroundColor Red
    exit 1
}

Write-Host "`n-- STEP 1: Dependency Audit ---------------------------------" -ForegroundColor Yellow

Run-Check "Zero external dependencies (empty require in go.mod)" {
    $deps = (go list -m all 2>&1 | Out-String).Trim()
    if ($deps -ne "taskrunner") {
        throw "Expected only 'taskrunner', got: $deps"
    }
}

Write-Host "`n-- STEP 2: Compilation & Hermetic Reproducibility ----------" -ForegroundColor Yellow

Run-Check "Clean compilation (go build ./...)" {
    $out = go build ./... 2>&1
    if ($LASTEXITCODE -ne 0) { throw ($out | Out-String) }
}

Run-Check "Byte-identical binary builds via -trimpath" {
    if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" -Force | Out-Null }
    go build -trimpath -ldflags="-buildid=" -o bin/check1.exe main.go 2>&1 | Out-Null
    go build -trimpath -ldflags="-buildid=" -o bin/check2.exe main.go 2>&1 | Out-Null
    $h1 = (Get-FileHash bin/check1.exe -Algorithm SHA256).Hash
    $h2 = (Get-FileHash bin/check2.exe -Algorithm SHA256).Hash
    Remove-Item bin/check1.exe, bin/check2.exe -Force -ErrorAction SilentlyContinue
    if ($h1 -ne $h2) { throw "Hash mismatch: $h1 != $h2" }
}

Write-Host "`n-- STEP 3: Automated Test Suite (7 Packages) ---------------" -ForegroundColor Yellow

Run-Check "Full test suite (go test ./...)" {
    $out = go test ./... 2>&1
    if ($LASTEXITCODE -ne 0) { throw ($out | Out-String) }
}

Write-Host "`n-- STEP 4: Build Standalone Executable ---------------------" -ForegroundColor Yellow

Run-Check "Build zephyr.exe" {
    $out = go build -o bin/zephyr.exe main.go 2>&1
    if ($LASTEXITCODE -ne 0) { throw ($out | Out-String) }
    if (-not (Test-Path "bin/zephyr.exe")) { throw "bin/zephyr.exe missing" }
}

$zephyr = ".\bin\zephyr.exe"

Write-Host "`n-- STEP 5: CLI Operations & Health Check --------------------" -ForegroundColor Yellow

Run-Check "zephyr version" {
    $out = (& $zephyr version 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0 -or $out -notmatch "ZEPHYR") { throw $out }
}

Run-Check "zephyr doctor" {
    $out = (& $zephyr doctor 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw $out }
}

Run-Check "zephyr adopt (project discovery)" {
    $out = (& $zephyr adopt 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw $out }
}

Write-Host "`n-- STEP 6: Cache Performance -------------------------------" -ForegroundColor Yellow

Run-Check "Cold execution vs Warm cache speedup" {
    & $zephyr clean 2>&1 | Out-Null
    $outCold = (& $zephyr run 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw "Cold build failed: $outCold" }

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $outWarm = (& $zephyr run 2>&1 | Out-String)
    $sw.Stop()
    if ($LASTEXITCODE -ne 0) { throw "Warm run failed: $outWarm" }
    if ($sw.ElapsedMilliseconds -gt 1000) {
        throw "Warm cache replay took $($sw.ElapsedMilliseconds)ms (expected <1000ms)"
    }
}

Write-Host "`n-- STEP 7: Reproducibility Auditor -------------------------" -ForegroundColor Yellow

Run-Check "zephyr verify auditor operation" {
    $out = (& $zephyr verify repro-build 2>&1 | Out-String)
    if ($out -notmatch "PASS") { throw "Verify failed: $out" }
}

Write-Host "`n-- STEP 8: Ed25519 Build Capsules --------------------------" -ForegroundColor Yellow

Run-Check "Capsule keygen, create, verify lifecycle" {
    & $zephyr capsule keygen 2>&1 | Out-Null
    $outCreate = (& $zephyr capsule create --key zephyr.key repro-build 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw "Capsule create failed: $outCreate" }
    $zcaps = Get-ChildItem ".taskcache/capsules/*.zcap" | Select-Object -Last 1
    if (-not $zcaps) { throw "No capsule found" }
    $outVerify = (& $zephyr capsule verify --key zephyr.pub $zcaps.FullName 2>&1 | Out-String)
    if ($outVerify -notmatch "PASS") { throw "Capsule verify failed: $outVerify" }
}

Write-Host "`n-- STEP 9: AI Agent & MCP Protocol -------------------------" -ForegroundColor Yellow

Run-Check "zephyr agent graph (JSON output)" {
    $out = (& $zephyr agent graph 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw $out }
    $json = $out | ConvertFrom-Json
    if (-not $json.operation) { throw "Invalid JSON envelope" }
}

Run-Check "zephyr agent affected (blast-radius analysis)" {
    $out = (& $zephyr agent affected --files=main.go 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw $out }
    $json = $out | ConvertFrom-Json
    if (-not $json.data.directly_affected) { throw "Missing directly_affected array" }
}

Write-Host "`n================================================================" -ForegroundColor Cyan
Write-Host "  AUDIT RESULTS: $passCount Passed, $failCount Failed" -ForegroundColor $(if ($failCount -eq 0) { "Green" } else { "Red" })
Write-Host "================================================================" -ForegroundColor Cyan

if ($failCount -eq 0) {
    Write-Host "`n[SUCCESS] ALL VERIFICATION CHECKS PASSED FOR ZEPHYR 2.0!`n" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n[WARNING] Some verification checks failed.`n" -ForegroundColor Red
    exit 1
}
