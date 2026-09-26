Write-Host "Starting Social Platform Local Development..." -ForegroundColor Cyan

# Check dependencies
if (!(Get-Command pnpm -ErrorAction SilentlyContinue)) {
    Write-Error "pnpm is required but not found in PATH."
    exit 1
}

Write-Host "1. Building internal packages..." -ForegroundColor Green
pnpm -r --filter=./packages/* build

Write-Host "2. Launching Next.js frontend..." -ForegroundColor Green
pnpm --filter web dev
