# install.ps1 — install aimap from the latest GitHub release (Windows)
# Usage: irm https://raw.githubusercontent.com/pgulb/aimap/main/install.ps1 | iex

$repo = "pgulb/aimap"
$binary = "aimap.exe"

# Detect architecture.
$arch = "amd64"
if ([Environment]::Is64BitOperatingSystem -eq $false) {
    Write-Host "32-bit systems are not supported."
    exit 1
}

$archive = "aimap_windows_${arch}.zip"

# Fetch latest release tag.
Write-Host "Fetching latest release..."
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
    $tag = $release.tag_name
} catch {
    Write-Host "Failed to determine latest release: $_"
    exit 1
}

Write-Host "Latest release: $tag"

$downloadUrl = "https://github.com/$repo/releases/download/$tag/$archive"

# Download to temp directory.
$tmpDir = Join-Path $env:TEMP "aimap_install_$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

$zipPath = Join-Path $tmpDir $archive
Write-Host "Downloading $archive..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -ErrorAction Stop
} catch {
    Write-Host "Download failed: $_"
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
    exit 1
}

Write-Host "Extracting..."
Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

$exePath = Join-Path $tmpDir $binary

# Determine install location.
$destDir = $null
$candidates = @(
    [Environment]::GetFolderPath("System")        # C:\Windows\System32
    (Join-Path $env:LOCALAPPDATA "Programs")       # ~\AppData\Local\Programs
    (Join-Path $env:USERPROFILE ".local\bin")      # ~\.local\bin
)

foreach ($dir in $candidates) {
    $testDir = [Environment]::ExpandEnvironmentVariables($dir)
    if (Test-Path $testDir) {
        try {
            $testFile = Join-Path $testDir "aimap_test_write"
            [System.IO.File]::WriteAllText($testFile, "")
            [System.IO.File]::Delete($testFile)
            $destDir = $testDir
            break
        } catch {
            continue
        }
    }
}

if (-not $destDir) {
    $destDir = Join-Path $env:USERPROFILE ".local\bin"
    New-Item -ItemType Directory -Path $destDir -Force | Out-Null
}

$destPath = Join-Path $destDir $binary
Copy-Item -Path $exePath -Destination $destPath -Force

Write-Host ""
Write-Host "Installed to $destPath"

# Check if destination is on PATH.
$paths = $env:PATH -split ";"
if ($paths -contains $destDir) {
    Write-Host "  (already in PATH)"
} else {
    Write-Host ""
    Write-Host "Note: $destDir is not in your PATH."
    Write-Host "Add it manually or restart your terminal if it was recently added."
}

# Cleanup.
Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Installation complete. Run 'aimap' to get started."
