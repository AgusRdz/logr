$ErrorActionPreference = "Stop"

$Repo = "AgusRdz/logr"
$InstallDir = if ($env:LOGR_INSTALL_DIR) { $env:LOGR_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\logr" }

$Arch = if (
    [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq `
    [System.Runtime.InteropServices.Architecture]::Arm64
) { "arm64" } else { "amd64" }

$Binary = "logr-windows-$Arch.exe"

if (-not $env:LOGR_VERSION) {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $env:LOGR_VERSION = $Release.tag_name
}

if (-not $env:LOGR_VERSION) {
    Write-Error "error: failed to determine latest version"
    exit 1
}

$BaseUrl = "https://github.com/$Repo/releases/download/$($env:LOGR_VERSION)"
Write-Host "installing logr $($env:LOGR_VERSION) (windows/$Arch)..."

# Download to temp dir and verify checksum before installing
$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null

try {
    $TmpBinary   = Join-Path $TmpDir $Binary
    $TmpChecksums = Join-Path $TmpDir "checksums.txt"

    Invoke-WebRequest -Uri "$BaseUrl/$Binary"        -OutFile $TmpBinary
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt"  -OutFile $TmpChecksums

    $Expected = (Get-Content $TmpChecksums | Where-Object { $_ -match [regex]::Escape($Binary) }) -split '\s+' | Select-Object -First 1
    $Actual   = (Get-FileHash -Algorithm SHA256 $TmpBinary).Hash.ToLower()

    if (-not $Expected -or $Expected -ne $Actual) {
        Write-Error "error: checksum mismatch - aborting install"
        exit 1
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $Destination = Join-Path $InstallDir "logr.exe"
    Move-Item -Force $TmpBinary $Destination
    Write-Host "installed logr to $Destination"
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}

# Add to user PATH if not already present
$UserPath   = [Environment]::GetEnvironmentVariable("PATH", "User")
$CleanDir   = $InstallDir.TrimEnd("\")
$PathParts  = $UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }

if ($PathParts -notcontains $CleanDir) {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "added $InstallDir to PATH"
}
$env:PATH = "$InstallDir;$env:PATH"

# Broadcast PATH change so open terminals pick it up without restart
$HWND_BROADCAST = [IntPtr]0xffff
$WM_SETTINGCHANGE = 0x001a
if (-not ("Win32.User32" -as [type])) {
    $MethodDef = @'
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(
    IntPtr hWnd, uint Msg, IntPtr wParam, string lParam,
    uint fuFlags, uint uTimeout, out IntPtr lpdwResult);
'@
    Add-Type -MemberDefinition $MethodDef -Name "User32" -Namespace "Win32" | Out-Null
}
$result = [IntPtr]::Zero
[Win32.User32]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [IntPtr]::Zero, "Environment", 2, 100, [ref]$result) | Out-Null

Write-Host ""
Write-Host "quick start:"
Write-Host "  Get-Content app.log | logr"
Write-Host "  logr app.log --level error --since 30m"
Write-Host "  logr --follow app.log --level error"
