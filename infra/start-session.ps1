#Requires -Version 5.1
<#
.SYNOPSIS
  Sobe o essencial da demo AWS: apply, publish API, health, publish front.

.DESCRIPTION
  Nao usa Docker Compose. Ordem do runbook P1:
  1) Valida credenciais do profile AWS default
  2) terraform init (se preciso) + apply
  3) publish-api.ps1
  4) espera GET {api_url}/health = 200
  5) publish-frontend.ps1

  Docker Desktop so e necessario para o build/push da imagem da API.

.EXAMPLE
  cd infra
  .\start-session.ps1
.EXAMPLE
  .\start-session.ps1 -Yes
#>
param(
    [switch]$Yes,
    [switch]$SkipPublish
)

$ErrorActionPreference = "Stop"
$InfraDir = $PSScriptRoot

function Write-Step([string]$msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Ok([string]$msg) { Write-Host "OK: $msg" -ForegroundColor Green }

function Assert-AwsDefault {
    if ($env:AWS_PROFILE) {
        Write-Host "Limpando AWS_PROFILE=$($env:AWS_PROFILE) (script usa so o profile default)." -ForegroundColor Yellow
        Remove-Item Env:AWS_PROFILE
    }
    $env:AWS_DEFAULT_REGION = "us-east-1"

    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    cmd /c "aws sts get-caller-identity --profile default --region us-east-1 >nul 2>&1"
    $ok = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $prev
    if ($ok) { return }

    $credPath = Join-Path $env:USERPROFILE ".aws\credentials"
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $key = (aws configure get aws_access_key_id --profile default 2>$null)
    $ErrorActionPreference = $prev

    if (-not $key) {
        $hint = "aws_access_key_id vazio"
    }
    elseif ($key -notmatch '^AKIA') {
        $prefix = $key.Substring(0, [Math]::Min(4, $key.Length))
        $hint = "key comeca com '$prefix' (precisa ser AKIA...)"
    }
    else {
        $hint = "key AKIA rejeitada pela AWS (crie uma nova no console)"
    }

    Write-Host ""
    Write-Host "Credenciais default invalidas: $hint" -ForegroundColor Red
    Write-Host "Edite: $credPath"
    Write-Host "Bloco [default] precisa de aws_access_key_id=AKIA... e secret correto."
    Write-Host "Region em .aws\config: us-east-1"
    Write-Host "Teste: aws sts get-caller-identity --profile default --region us-east-1"
    throw "Credenciais AWS default invalidas."
}

function Assert-Tools {
    foreach ($cmd in @("terraform", "aws", "docker")) {
        if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
            throw "Comando '$cmd' nao encontrado no PATH."
        }
    }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    cmd /c "docker info >nul 2>&1"
    $dockerOk = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $prevEap
    if (-not $dockerOk) {
        throw "Docker Desktop precisa estar aberto (build/push da imagem da API)."
    }
}

function Ensure-Tfvars {
    $example = Join-Path $InfraDir "terraform.tfvars.example"
    $tfvars = Join-Path $InfraDir "terraform.tfvars"
    if (-not (Test-Path $tfvars)) {
        if (-not (Test-Path $example)) { throw "Falta terraform.tfvars.example em infra/" }
        Copy-Item $example $tfvars
        Write-Host "Criado infra/terraform.tfvars a partir do example."
    }
}

function Wait-Http200([string]$url, [int]$timeoutMinutes) {
    Write-Step "Aguardando health: $url"
    $deadline = (Get-Date).AddMinutes($timeoutMinutes)
    while ((Get-Date) -lt $deadline) {
        try {
            $resp = Invoke-WebRequest -Uri $url -Method GET -UseBasicParsing -TimeoutSec 15
            if ($resp.StatusCode -eq 200) {
                Write-Ok "HTTP 200"
                return
            }
        }
        catch {
            Start-Sleep -Seconds 8
        }
        Start-Sleep -Seconds 4
    }
    throw "Timeout: $url nao retornou 200 em ${timeoutMinutes} min."
}

Assert-Tools
Assert-AwsDefault
Write-Ok "AWS default OK"

Write-Host ""
Write-Host "EstudaJa - sessao AWS (apresentacao)"
Write-Host "  Stack ligada = custo. Depois: .\stop-session.ps1"
Write-Host "  Budget em infra/budget/ NAO e tocado."
Write-Host ""

if (-not $Yes) {
    $ans = Read-Host "Rodar terraform apply + publish agora? [S/n]"
    if ($ans -and $ans -notmatch '^[sSyY]') {
        Write-Host "Cancelado."
        exit 0
    }
}

Ensure-Tfvars

Push-Location $InfraDir
try {
    if (-not (Test-Path (Join-Path $InfraDir ".terraform"))) {
        Write-Step "terraform init"
        terraform init
        if ($LASTEXITCODE -ne 0) { throw "terraform init falhou" }
    }

    Write-Step "terraform apply"
    terraform apply -auto-approve
    if ($LASTEXITCODE -ne 0) { throw "terraform apply falhou" }

    $frontendUrl = terraform output -raw frontend_url
    $apiUrl = (terraform output -raw api_url).TrimEnd("/")

    Write-Ok "Infra criada"
    Write-Host "  frontend_url: $frontendUrl"
    Write-Host "  api_url:      $apiUrl"

    if ($SkipPublish) {
        Write-Host "SkipPublish: rode .\publish-api.ps1 e .\publish-frontend.ps1"
        exit 0
    }

    Write-Step "Publish API (build linux/arm64 -> ECR -> ECS)"
    & (Join-Path $InfraDir "publish-api.ps1")
    if ($LASTEXITCODE -ne 0) { throw "publish-api.ps1 falhou" }

    Wait-Http200 "$apiUrl/health" 10

    Write-Step "Publish frontend (VITE_API_URL -> S3 -> CloudFront)"
    & (Join-Path $InfraDir "publish-frontend.ps1")
    if ($LASTEXITCODE -ne 0) { throw "publish-frontend.ps1 falhou" }

    Write-Host ""
    Write-Ok "Demo pronta"
    Write-Host "  Abra: $frontendUrl"
    Write-Host "  Logins: admin@estudaja.com / admin123 (tambem professor/aluno)"
    Write-Host "  Encerrar: .\stop-session.ps1"
}
finally {
    Pop-Location
}
