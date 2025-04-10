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

function Test-ValidProjectName {
    param (
        [string]$ProjectName
    )
    
    if ($ProjectName -match "\s") {
        return $false, "Project name cannot contain spaces"
    }
    
    if ($ProjectName -cmatch "[A-Z]") {
        return $false, "Project name cannot contain uppercase letters"
    }
    
    return $true, ""
}

function Update-ImportPaths {
    param (
        [string]$OldName,
        [string]$NewName,
        [string]$Directory
    )
    
    Write-Styled "Updating import paths in source files..." -Color $Theme.Primary -Prefix "Config"
    
    try {
        $goFiles = Get-ChildItem -Path $Directory -Filter "*.go" -Recurse
        $totalCount = 0
        
        foreach ($file in $goFiles) {
            $content = Get-Content -Path $file.FullName -Raw
            $originalContent = $content
            
            $updatedContent = $content -replace "([`"'])$OldName(/[^`"']*)?([`"'])", "`$1$NewName`$2`$3"
            
     
            if ($updatedContent -ne $originalContent) {
                Set-Content -Path $file.FullName -Value $updatedContent
                $totalCount++
            }
        }
        
        if ($totalCount -gt 0) {
            Write-Styled "Updated import paths in $totalCount files" -Color $Theme.Success -Prefix "Config"
        } else {
            Write-Styled "No import paths needed updating" -Color $Theme.Info -Prefix "Config"
        }
    }
    catch {
        Write-Styled "Some import paths may not have been updated properly" -Color $Theme.Warning -Prefix "Config"
        Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
    }
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
    
    $projectName = ""
    $validName = $false
    
    while (-not $validName) {
        $projectName = Read-Host "Enter project name (default: $defaultName)"
        if ([string]::IsNullOrWhiteSpace($projectName)) {
            $projectName = $defaultName
        }
        
        $validationResult = Test-ValidProjectName -ProjectName $projectName
        $validName = $validationResult[0]
        
        if (-not $validName) {
            Write-Styled $validationResult[1] -Color $Theme.Error -Prefix "Error"
            Write-Styled "Please enter a valid project name (lowercase letters, no spaces)" -Color $Theme.Warning
        }
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
    
    Write-Styled "Cleaning up unnecessary files..." -Color $Theme.Primary -Prefix "Files"
    Push-Location $projectDir
    
    try {
        $filesToRemove = @("img", "README.md", "MANIFEST.md", "CODE_OF_CONDUCT.md", "ROADMAP.md", "SECURITY.md")
        foreach ($file in $filesToRemove) {
            if (Test-Path $file) {
                Remove-Item -Path $file -Recurse -Force
            }
        }
        
        $keepDocs = Read-Host "Do you want to keep local documentation? (Y/n)"
        if ($keepDocs -eq "n" -or $keepDocs -eq "N") {
            if (Test-Path "docs") {
                Remove-Item -Path "docs" -Recurse -Force
                Write-Styled "Documentation removed" -Color $Theme.Success -Prefix "Files"
            }
        } else {
            Write-Styled "Documentation kept" -Color $Theme.Success -Prefix "Files"
        }
        
        if (Test-Path "scripts") {
            Remove-Item -Path "scripts" -Recurse -Force
        }
        
        Write-Styled "Cleanup completed" -Color $Theme.Success -Prefix "Files"
    }
    catch {
        Write-Styled $_.Exception.Message -Color $Theme.Error -Prefix "Error"
    }
    
    Write-Styled "Initializing new Git repository..." -Color $Theme.Primary -Prefix "Git"
    
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
        
        Update-ImportPaths -OldName "goonairplanes" -NewName $projectName -Directory $projectDir
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