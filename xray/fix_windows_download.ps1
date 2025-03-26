# Xray Windows version download fix script (PowerShell)

# Configuration
$Version = "v1.8.23"  # Target version, modify as needed
$DownloadUrl = "https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/${Version}/Xray-windows-64.zip"
$DownloadDir = "xray\downloads"
$InstallDir = "xray\bin\${Version}"

Write-Host "=== Xray Windows Version Download Fix Script ===" -ForegroundColor Green
Write-Host "Version: $Version"
Write-Host "Download URL: $DownloadUrl"
Write-Host "Download Directory: $DownloadDir"
Write-Host "Install Directory: $InstallDir"

# Create necessary directories
if (-not (Test-Path $DownloadDir)) {
    New-Item -ItemType Directory -Path $DownloadDir -Force | Out-Null
    Write-Host "Created download directory: $DownloadDir"
}

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "Created install directory: $InstallDir"
}

# Check if executable already exists
$ExePath = Join-Path $InstallDir "xray.exe"
if (Test-Path $ExePath) {
    Write-Host "Xray executable already exists: $ExePath" -ForegroundColor Yellow
    $fileSize = (Get-Item $ExePath).Length
    Write-Host "File size: $fileSize bytes"
    
    $choice = Read-Host "Download again? (y/n)"
    if ($choice -ne "y" -and $choice -ne "Y") {
        Write-Host "Operation cancelled." -ForegroundColor Yellow
        exit 0
    }
}

# Download Xray
$ZipFile = Join-Path $DownloadDir "Xray-windows-64.zip"
Write-Host "Downloading Xray Windows version..." -ForegroundColor Cyan

try {
    # Create temporary progress bar
    for ($i = 0; $i -le 100; $i += 10) {
        Write-Progress -Activity "Downloading Xray..." -Status "$i% Complete" -PercentComplete $i
        Start-Sleep -Milliseconds 100
    }
    
    # Actually download
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipFile -UseBasicParsing
    Write-Progress -Activity "Downloading Xray..." -Status "Complete" -PercentComplete 100 -Completed
}
catch {
    Write-Host "Download failed:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}

# Verify download
if (-not (Test-Path $ZipFile)) {
    Write-Host "Download failed! File does not exist" -ForegroundColor Red
    exit 1
}

$fileSize = (Get-Item $ZipFile).Length
Write-Host "Download complete, file size: $fileSize bytes"

if ($fileSize -eq 0) {
    Write-Host "Error: Downloaded file is empty!" -ForegroundColor Red
    exit 1
}

# Use Node.js tool to extract, if available
if (Test-Path "tools\download_xray.js") {
    Write-Host "Using Node.js tool to extract..." -ForegroundColor Cyan
    try {
        Push-Location tools
        if (-not (Test-Path "node_modules\extract-zip\index.js")) {
            Write-Host "Installing dependencies..."
            npm install
        }
        Write-Host "Running Node.js Xray download tool..."
        node download_xray.js $Version
        Pop-Location
    }
    catch {
        Write-Host "Node.js tool failed:" -ForegroundColor Red
        Write-Host $_.Exception.Message
        Write-Host "Trying PowerShell extraction..." -ForegroundColor Yellow
        
        # Use PowerShell built-in method to extract
        try {
            Expand-Archive -Path $ZipFile -DestinationPath $InstallDir -Force
        }
        catch {
            Write-Host "PowerShell extraction failed:" -ForegroundColor Red
            Write-Host $_.Exception.Message
            exit 1
        }
    }
}
else {
    Write-Host "Node.js tool not found, using PowerShell extraction..." -ForegroundColor Yellow
    try {
        Expand-Archive -Path $ZipFile -DestinationPath $InstallDir -Force
    }
    catch {
        Write-Host "PowerShell extraction failed:" -ForegroundColor Red
        Write-Host $_.Exception.Message
        exit 1
    }
}

# Check extracted files
if (Test-Path $ExePath) {
    $fileSize = (Get-Item $ExePath).Length
    Write-Host "Installation successful! Xray executable located at: $ExePath" -ForegroundColor Green
    Write-Host "File size: $fileSize bytes"
}
else {
    Write-Host "Installation failed! Xray executable not found" -ForegroundColor Red
    exit 1
}

Write-Host "Complete!" -ForegroundColor Green 