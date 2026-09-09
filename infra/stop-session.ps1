#Requires -Version 5.1
<#
.SYNOPSIS
  Encerra a demo AWS: terraform destroy (corta custo). Nao toca em infra/budget/.

.EXAMPLE
  cd infra
  .\stop-session.ps1
.EXAMPLE
  .\stop-session.ps1 -Yes
#>
param(
    [switch]$Yes
)

$ErrorActionPreference = "Stop"
$InfraDir = $PSScriptRoot

function Write-Step([string]$msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Ok([string]$msg) { Write-Host "OK: $msg" -ForegroundColor Green }

function Assert-AwsDefault {
    if ($env:AWS_PROFILE) { Remove-Item Env:AWS_PROFILE }
    $env:AWS_DEFAULT_REGION = "us-east-1"
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    cmd /c "aws sts get-caller-identity --profile default --region us-east-1 >nul 2>&1"
    $ok = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $prev
    if (-not $ok) {
        throw "Credenciais default invalidas. Corrija ~/.aws/credentials (Access Key AKIA...)."
    }
}

if (-not (Get-Command terraform -ErrorAction SilentlyContinue)) {
    throw "terraform nao encontrado no PATH."
}
if (-not (Get-Command aws -ErrorAction SilentlyContinue)) {
    throw "aws CLI nao encontrado no PATH."
}

Assert-AwsDefault
Write-Ok "AWS default OK"

Write-Host ""
Write-Host "EstudaJa - destroy da demo AWS"
Write-Host "  Remove ECS/ALB/RDS/S3/CloudFront/IVS da sessao."
Write-Host "  NAO destroi o budget (infra/budget/)."
Write-Host ""

if (-not $Yes) {
    $ans = Read-Host "Confirmar terraform destroy? [S/n]"
    if ($ans -and $ans -notmatch '^[sSyY]') {
        Write-Host "Cancelado."
        exit 0
    }
}

Push-Location $InfraDir
try {
    if (-not (Test-Path (Join-Path $InfraDir ".terraform"))) {
        Write-Step "terraform init"
        terraform init
        if ($LASTEXITCODE -ne 0) { throw "terraform init falhou" }
    }

    Write-Step "terraform destroy"
    terraform destroy -auto-approve
    if ($LASTEXITCODE -ne 0) { throw "terraform destroy falhou" }

    Write-Ok "Stack destruida - custo da demo deve ir a ~0 (exceto budget)."
}
finally {
    Pop-Location
}
