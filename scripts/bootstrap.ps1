Write-Host "Bootstrapping Career / LinkedIn / Social Platform..." -ForegroundColor Cyan

# 1. Check prerequisites
Write-Host "Checking prerequisites..." -ForegroundColor Yellow
if (!(Get-Command pnpm -ErrorAction SilentlyContinue)) {
    Write-Error "pnpm is not installed. Please install pnpm."
    exit 1
}

# 2. Setup .env file
if (!(Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env from .env.example" -ForegroundColor Green
} else {
    Write-Host ".env already exists." -ForegroundColor DarkGray
}

# 3. Install packages
Write-Host "Installing workspace dependencies..." -ForegroundColor Yellow
pnpm install

# 4. Build packages
Write-Host "Building initial packages..." -ForegroundColor Yellow
pnpm -r --filter=./packages/* build

# 5. Build Go tools
Write-Host "Building Go core binaries..." -ForegroundColor Yellow
& "C:\Program Files\Go\bin\go.exe" build -C services/core ./cmd/...

Write-Host "Bootstrap completed successfully!" -ForegroundColor Green