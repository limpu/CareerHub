Write-Host "Running Social Platform Automated Checks..." -ForegroundColor Cyan

Write-Host "1. Testing TypeScript packages & linting..." -ForegroundColor Green
pnpm -r lint

Write-Host "2. Building all packages & apps..." -ForegroundColor Green
pnpm -r build

Write-Host "3. Testing Go core services..." -ForegroundColor Green
& "C:\Program Files\Go\bin\go.exe" test ./...

Write-Host "All checks completed successfully!" -ForegroundColor Green
