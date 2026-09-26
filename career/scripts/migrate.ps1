param(
    [string]$Command = "validate",
    [string]$MigrationsDir = "../../migrations"
)

Write-Host "Running database migration tool ($Command)..." -ForegroundColor Cyan

$goBinary = "go"
if (Test-Path "C:\Program Files\Go\bin\go.exe") {
    $goBinary = "C:\Program Files\Go\bin\go.exe"
}

& $goBinary run -C services/core ./cmd/migrate -dir $MigrationsDir -cmd $Command
if ($LASTEXITCODE -ne 0) {
    Write-Error "Migration command failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}