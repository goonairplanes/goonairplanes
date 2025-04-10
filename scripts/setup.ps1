$Theme = @{
    Primary   = 'Cyan'
    Success   = 'Green'
    Warning   = 'Yellow'
    Error     = 'Red'
    Info      = 'White'
}

$Logo = @"
   ██████╗  ██████╗      ██████╗ ███╗   ██╗     █████╗ ██╗██████╗ ██████╗ ██╗      █████╗ ███╗   ██╗███████╗███████╗
  ██╔════╝ ██╔═══██╗    ██╔═══██╗████╗  ██║    ██╔══██╗██║██╔══██╗██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝██╔════╝
  ██║  ███╗██║   ██║    ██║   ██║██╔██╗ ██║    ███████║██║██████╔╝██████╔╝██║     ███████║██╔██╗ ██║█████╗  ███████╗
  ██║   ██║██║   ██║    ██║   ██║██║╚██╗██║    ██╔══██║██║██╔══██╗██╔═══╝ ██║     ██╔══██║██║╚██╗██║██╔══╝  ╚════██║
  ╚██████╔╝╚██████╔╝    ╚██████╔╝██║ ╚████║    ██║  ██║██║██║  ██║██║     ███████╗██║  ██║██║ ╚████║███████╗███████║
   ╚═════╝  ╚═════╝      ╚═════╝ ╚═╝  ╚═══╝    ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝
"@

function Write-Styled {
    param (
        [string]$Message,
        [string]$Color = $Theme.Info,
        [string]$Prefix = "",
        [switch]$NoNewline
    )
    $symbol = switch ($Color) {
        $Theme.Success { "[✓]" }
        $Theme.Error   { "[✗]" }
        $Theme.Warning { "[!]" }
        default        { "[*]" }
    }
    
    $output = if ($Prefix) { "$symbol $Prefix :: $Message" } else { "$symbol $Message" }
    if ($NoNewline) {
        Write-Host $output -ForegroundColor $Color -NoNewline
    } else {
        Write-Host $output -ForegroundColor $Color
    }
}

function Test-CommandExists {
    param (
        [string]$Command
    )
    
    $exists = $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
    return $exists
}

Write-Host $Logo -ForegroundColor $Theme.Primary
Write-Host "Go on Airplanes - Setup Wizard" -ForegroundColor $Theme.Primary
Write-Host "Fly high with simple web development`n" -ForegroundColor $Theme.Info

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Setup-GoOnAirplanes {
    Write-Styled "Checking for Git installation..." -Color $Theme.Primary -Prefix "System"
    if (-not (Test-CommandExists "git")) {
        Write-Styled "Git is not installed on your system." -Color $Theme.Error -Prefix "Error"
        Write-Styled "Please install Git from https://git-scm.com/downloads" -Color $Theme.Info
        return $false
    }
    Write-Styled "Git is installed" -Color $Theme.Success -Prefix "System"
    
    Write-Styled "Checking for Go installation..." -Color $Theme.Primary -Prefix "System"
    if (-not (Test-CommandExists "go")) {
        Write-Styled "Go is not installed on your system." -Color $Theme.Error -Prefix "Error"
        Write-Styled "Please install Go from https://golang.org/dl/" -Color $Theme.Info
        return $false
    }
    
    $goVersion = (go version) -replace "go version go([0-9]+\.[0-9]+\.[0-9]+).*", '$1'
    Write-Styled "Go $goVersion is installed" -Color $Theme.Success -Prefix "System"
    
    $defaultName = Split-Path -Path (Get-Location) -Leaf
    Write-Host "`nProject Setup" -ForegroundColor $Theme.Primary
    
    $projectName = Read-Host "Enter project name (default: $defaultName)"
    if ([string]::IsNullOrWhiteSpace($projectName)) {
        $projectName = $defaultName
    }
    
    $useCurrentDir = Read-Host "Use current directory? (Y/n)"
    if ($useCurrentDir -eq "" -or $useCurrentDir -eq "y" -or $useCurrentDir -eq "Y") {
        $projectDir = Get-Location
        $inCurrentDir = $true
    } else {
        $projectDir = Join-Path (Get-Location) $projectName
        $inCurrentDir = $false
        
        if (-not (Test-Path $projectDir)) {
            New-Item -ItemType Directory -Path $projectDir | Out-Null
        }
    }
    
    Write-Styled "Cloning Go on Airplanes repository..." -Color $Theme.Primary -Prefix "Git"
    
    if ($inCurrentDir) {
        $tempDir = Join-Path $env:TEMP "goa-temp-$(Get-Random)"
        New-Item -ItemType Directory -Path $tempDir | Out-Null
        
        try {
            Push-Location $tempDir
            git clone https://github.com/kleeedolinux/goonairplanes.git . 
            
            Get-ChildItem -Force | Where-Object { $_.Name -ne ".git" } | Copy-Item -Destination $projectDir -Recurse -Force
            
            Pop-Location
            Remove-Item -Path $tempDir -Recurse -Force
        }
        catch {
            Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
            Pop-Location
            return $false
        }
    }
    else {
        try {
            git clone https://github.com/kleeedolinux/goonairplanes.git $projectDir 
            Remove-Item -Path (Join-Path $projectDir ".git") -Recurse -Force
        }
        catch {
            Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
            return $false
        }
    }
    
    Write-Styled "Repository cloned successfully" -Color $Theme.Success -Prefix "Git"
    
    Write-Styled "Initializing new Git repository..." -Color $Theme.Primary -Prefix "Git"
    Push-Location $projectDir
    
    try {
        git init 
        git add . 
        git commit -m "Initial commit: Go on Airplanes project" 
        
        Write-Styled "Git repository initialized" -Color $Theme.Success -Prefix "Git"
    }
    catch {
        Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
    }
    
    Write-Styled "Updating project configuration..." -Color $Theme.Primary -Prefix "Config"
    
    $goModPath = Join-Path $projectDir "go.mod"
    if (Test-Path $goModPath) {
        $goMod = Get-Content $goModPath -Raw
        $goMod = $goMod -replace "module goonairplanes", "module $projectName"
        Set-Content -Path $goModPath -Value $goMod
        Write-Styled "Updated module name in go.mod" -Color $Theme.Success -Prefix "Config"
    }
    
    Write-Styled "Installing dependencies..." -Color $Theme.Primary -Prefix "Go"
    try {
        go mod tidy
        Write-Styled "Dependencies installed successfully" -Color $Theme.Success -Prefix "Go"
    }
    catch {
        Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
    }
    
    Write-Styled "`nSetup completed successfully!" -Color $Theme.Success -Prefix "Done"
    Write-Styled "Your Go on Airplanes project is ready at: $projectDir" -Color $Theme.Info
    
    Write-Host "`nTo run your application:" -ForegroundColor $Theme.Primary
    Write-Host "  cd $projectDir" -ForegroundColor $Theme.Info
    Write-Host "  go run main.go" -ForegroundColor $Theme.Info
    
    Pop-Location
    return $true
}

try {
    $success = Setup-GoOnAirplanes
    if (-not $success) {
        Write-Styled "Setup failed" -Color $Theme.Error -Prefix "Error"
    }
}
catch {
    Write-Styled "Setup failed" -Color $Theme.Error -Prefix "Error"
    Write-Styled $_.Exception.Message -Color $Theme.Error
}
finally {
    Write-Host "`nPress any key to exit..." -ForegroundColor $Theme.Info
    $null = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
} 