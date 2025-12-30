# Build script for cross-platform compilation

$ErrorActionPreference = "Stop"

$VERSION = "1.0.0"
$BUILD_DIR = "build"
$PLATFORMS = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"},
    @{OS="windows"; ARCH="arm64"; EXT=".exe"},
    @{OS="linux"; ARCH="amd64"; EXT=""},
    @{OS="linux"; ARCH="arm64"; EXT=""},
    @{OS="darwin"; ARCH="amd64"; EXT=""},
    @{OS="darwin"; ARCH="arm64"; EXT=""}
)

Write-Host "Building Enterprise Agent v$VERSION" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green

# Create build directory
if (!(Test-Path $BUILD_DIR)) {
    New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null
}

# Build for each platform
foreach ($platform in $PLATFORMS) {
    $os = $platform.OS
    $arch = $platform.ARCH
    $ext = $platform.EXT
    
    $outputName = "agent-$os-$arch$ext"
    $outputPath = Join-Path $BUILD_DIR $outputName
    
    Write-Host "`nBuilding for $os/$arch..." -ForegroundColor Cyan
    
    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"
    
    $ldflags = "-s -w -X main.version=$VERSION"
    
    go build -ldflags $ldflags -o $outputPath ./cmd/agent
    
    if ($LASTEXITCODE -eq 0) {
        $size = (Get-Item $outputPath).Length / 1MB
        Write-Host "  ✓ Built $outputName ($([math]::Round($size, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Build failed for $os/$arch" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n✓ All builds completed successfully!" -ForegroundColor Green
Write-Host "`nBuild artifacts in: $BUILD_DIR" -ForegroundColor Yellow

# List all built binaries
Write-Host "`nBuilt binaries:" -ForegroundColor Cyan
Get-ChildItem $BUILD_DIR | ForEach-Object {
    $size = $_.Length / 1MB
    Write-Host "  - $($_.Name) ($([math]::Round($size, 2)) MB)"
}
