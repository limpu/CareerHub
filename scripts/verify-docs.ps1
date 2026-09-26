Write-Host "Running Documentation Verification Suite..." -ForegroundColor Cyan

$requiredDocs = @(
    "doc/README.md",
    "doc/CONTEXT.md",
    "doc/PLAN.md",
    "doc/TASKS.md",
    "doc/PROGRESS.md",
    "doc/AI-AGENT-ARCHITECTURE.md"
)

$failed = $false

foreach ($doc in $requiredDocs) {
    if (!(Test-Path $doc)) {
        Write-Error "Missing required documentation file: $doc"
        $failed = $true
        continue
    }

    $fileInfo = Get-Item $doc
    if ($fileInfo.Length -eq 0) {
        Write-Error "Documentation file is empty: $doc"
        $failed = $true
        continue
    }

    $lines = (Get-Content $doc).Count
    Write-Host "  [OK] $doc ($($fileInfo.Length) bytes, $lines lines)" -ForegroundColor Green
}

# Validate TASKS.md task IDs
$tasksContent = Get-Content "doc/TASKS.md" -Raw
$taskMatches = [regex]::Matches($tasksContent, '`([A-Z]{3}-[A-Z0-9-]+)`')
$taskIds = @()
foreach ($m in $taskMatches) {
    $taskIds += $m.Groups[1].Value
}
$uniqueTasks = $taskIds | Select-Object -Unique

Write-Host "  [OK] Scanned $($taskIds.Count) task references ($($uniqueTasks.Count) unique task IDs)" -ForegroundColor Green

if ($failed) {
    Write-Error "Documentation verification failed."
    exit 1
} else {
    Write-Host "All documentation checks passed successfully!" -ForegroundColor Green
}