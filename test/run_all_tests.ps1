# ZEPHYR 2.0 - MASTER STRESS AND ADVERSARIAL TEST HARNESS (ALL 14 SUITES)
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host "ZEPHYR 2.0 -- MASTER BRUTAL CHAOS AND STRESS TEST ORCHESTRATOR" -ForegroundColor Cyan
Write-Host "Author: Naman Swami | Runtime: 100% Go Stdlib | Target Directory: test/" -ForegroundColor Cyan
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host ""

$rootTestDir = $PSScriptRoot
$suites = @(
    @{ Name = "Suite 01: DAG and Topology Stress"; Dir = "$rootTestDir\01_dag_adversarial"; Script = "dag_test_runner.go" },
    @{ Name = "Suite 02: Security, Tar Slip and Anti-Tamper"; Dir = "$rootTestDir\02_security_jail"; Script = "security_tests.go" },
    @{ Name = "Suite 03: Polyglot Project Adoption"; Dir = "$rootTestDir\03_polyglot_adoption"; Script = "adopt_tests.go" },
    @{ Name = "Suite 04: CAS Storage and Caching Engine"; Dir = "$rootTestDir\04_caching_and_cas"; Script = "cas_tests.go" },
    @{ Name = "Suite 05: Concurrency, Timeouts and Retries"; Dir = "$rootTestDir\05_concurrency_and_faults"; Script = "fault_tests.go" },
    @{ Name = "Suite 06: AI Agent Interface and MCP Protocol"; Dir = "$rootTestDir\06_agent_and_mcp"; Script = "agent_mcp_tests.go" },
    @{ Name = "Suite 07: Reproducibility Auditor and Diffing"; Dir = "$rootTestDir\07_reproducibility_auditor"; Script = "repro_tests.go" },
    @{ Name = "Suite 08: Cache Concurrency Race Conditions"; Dir = "$rootTestDir\08_brutal_concurrency_race"; Script = "race_bench.go" },
    @{ Name = "Suite 09: Hardlink In-Place Mutation Trap"; Dir = "$rootTestDir\09_brutal_hardlink_mutation"; Script = "hardlink_mutation.go" },
    @{ Name = "Suite 10: Deep Process Tree Termination"; Dir = "$rootTestDir\10_brutal_zombie_process_tree"; Script = "zombie_tree_bench.go" },
    @{ Name = "Suite 11: Real-Time Stream Redaction Masking"; Dir = "$rootTestDir\11_brutal_stream_redaction_verify"; Script = "stream_redact_bench.go" },
    @{ Name = "Suite 12: Output Globbing in 'dist/' Dir"; Dir = "$rootTestDir\12_brutal_output_globbing_dist"; Script = "dist_glob_bench.go" },
    @{ Name = "Suite 13: Filesystem Edge Cases (Zero-Byte/Unicode)"; Dir = "$rootTestDir\13_brutal_filesystem_edge_cases"; Script = "fs_edge_bench.go" },
    @{ Name = "Suite 14: Monorepo Cross-Package Circularity"; Dir = "$rootTestDir\14_brutal_monorepo_circularity"; Script = "monorepo_bench.go" }
)

$suiteResults = @()
$totalPassed = 0
$totalFailed = 0
$globalStart = Get-Date

foreach ($s in $suites) {
    Write-Host ">>> Executing $($s.Name)..." -ForegroundColor Yellow
    Push-Location $s.Dir
    $start = Get-Date
    go run $s.Script
    $exitCode = $LASTEXITCODE
    $dur = (Get-Date) - $start
    Pop-Location

    $status = if ($exitCode -eq 0) { "PASS" } else { "FAIL" }
    if ($exitCode -eq 0) { $totalPassed++ } else { $totalFailed++ }

    $suiteResults += [PSCustomObject]@{
        Suite    = $s.Name
        Status   = $status
        Duration = "$([math]::Round($dur.TotalSeconds, 2))s"
        ExitCode = $exitCode
    }
    Write-Host ""
}

$globalDur = (Get-Date) - $globalStart

Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host "ZEPHYR 2.0 MASTER TEST SUITE EXECUTION SUMMARY (ALL 14 SUITES)" -ForegroundColor Cyan
Write-Host "==============================================================================" -ForegroundColor Cyan
$suiteResults | Format-Table -AutoSize

Write-Host "Total Suites: $($suites.Count) | Passed: $totalPassed | Failed: $totalFailed" -ForegroundColor $(if ($totalFailed -eq 0) { "Green" } else { "Red" })
Write-Host "Total Execution Time: $([math]::Round($globalDur.TotalSeconds, 2)) seconds" -ForegroundColor Cyan
Write-Host "==============================================================================" -ForegroundColor Cyan
