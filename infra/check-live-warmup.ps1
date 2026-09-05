#Requires -Version 5.1
<#
.SYNOPSIS
  Checagem mínima de warm-up de streaming (P2) — health da API + sinal IVS opcional.

.DESCRIPTION
  NÃO aplica nem destrói a stack 001. Só valida:
  1) GET {ApiUrl}/health = 200
  2) (opcional) aws ivs get-stream no canal — se ChannelArn informado e AWS CLI disponível

  Uso típico no T−10 / T−5 do runbook 003, depois do apply/publish manuais.

.EXAMPLE
  .\check-live-warmup.ps1 -ApiUrl "https://xxxx.cloudfront.net"
.EXAMPLE
  .\check-live-warmup.ps1 -ApiUrl "https://xxxx.cloudfront.net" -ChannelArn (terraform output -raw ivs_channel_arn)
#>
param(
  [Parameter(Mandatory = $true)]
  [string]$ApiUrl,

  [Parameter(Mandatory = $false)]
  [string]$ChannelArn = "",

  [Parameter(Mandatory = $false)]
  [string]$Region = "us-east-1"
)

$ErrorActionPreference = "Stop"

$base = $ApiUrl.TrimEnd("/")
$healthUrl = "$base/health"
Write-Host "1) Health API: GET $healthUrl"

try {
  $resp = Invoke-WebRequest -Uri $healthUrl -Method GET -UseBasicParsing -TimeoutSec 15
} catch {
  Write-Error "Health falhou: $_"
  exit 1
}

if ($resp.StatusCode -ne 200) {
  Write-Error "Health retornou HTTP $($resp.StatusCode) (esperado 200)."
  exit 1
}
Write-Host "   OK — HTTP 200"

if ([string]::IsNullOrWhiteSpace($ChannelArn)) {
  Write-Host "2) Sinal IVS: pulado (passe -ChannelArn para checar GetStream)."
  Write-Host "Warm-up streaming (health): OK. OBS/start live continuam manuais no checklist T−5."
  exit 0
}

Write-Host "2) Sinal IVS: GetStream em $ChannelArn ($Region)"
$aws = Get-Command aws -ErrorAction SilentlyContinue
if (-not $aws) {
  Write-Warning "AWS CLI não encontrado — não foi possível GetStream. Health já OK."
  exit 0
}

$streamJson = aws ivs get-stream --channel-arn $ChannelArn --region $Region --output json 2>&1
if ($LASTEXITCODE -ne 0) {
  Write-Host "   GetStream: sem stream ativo (normal se OBS ainda não estiver Live)."
  Write-Host "   Mensagem: $streamJson"
  Write-Host "Warm-up streaming (health): OK; sinal IVS ainda ausente — ligue OBS no T−5."
  exit 0
}

Write-Host "   OK — stream ativo:"
Write-Host $streamJson
Write-Host "Warm-up streaming (health + sinal): OK."
exit 0
