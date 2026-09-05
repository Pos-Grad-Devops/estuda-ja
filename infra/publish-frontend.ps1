# Build SPA com VITE_API_URL da sessão → S3 sync + invalidação CloudFront.
# Pré: API healthy (GET {api_url}/health = 200); Node ≥20 (nvm use 22 se necessário).
# Uso (a partir de infra/): .\publish-frontend.ps1

$ErrorActionPreference = "Stop"
$InfraDir = $PSScriptRoot
$RepoRoot = Split-Path -Parent $InfraDir
$FrontendDir = Join-Path $RepoRoot "frontend"

# Preferir Node 22 via nvm-windows quando o PATH default for antigo (Vite 8).
if (Get-Command nvm -ErrorAction SilentlyContinue) {
    Write-Host "==> nvm use 22.23.2"
    & nvm use 22.23.2 | Out-Null
    $nvmNode = "C:\nvm4w\nodejs"
    if (Test-Path (Join-Path $nvmNode "node.exe")) {
        # Colocar nvm na frente de "Program Files\nodejs" (Node 14 legado).
        $parts = $env:Path -split ";" | Where-Object {
            $_ -and ($_ -notmatch '(?i)\\nodejs\\?$') -and ($_ -notmatch '(?i)\\nvm4w\\nodejs')
        }
        $env:Path = (@($nvmNode) + $parts) -join ";"
    }
}

Write-Host "==> node $(node -v) / npm $(npm -v)"

Push-Location $InfraDir
try {
    $region = terraform output -raw aws_region
    $apiUrl = (terraform output -raw api_url).TrimEnd("/")
    $bucket = terraform output -raw s3_bucket_name
    $distId = terraform output -raw cloudfront_frontend_distribution_id
}
finally {
    Pop-Location
}

Write-Host "==> npm ci + build (VITE_API_URL=$apiUrl)"
Push-Location $FrontendDir
try {
    npm ci
    if ($LASTEXITCODE -ne 0) { throw "npm ci falhou" }
    $env:VITE_API_URL = $apiUrl
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "npm run build falhou" }
}
finally {
    Pop-Location
}

$dist = Join-Path $FrontendDir "dist"
Write-Host "==> s3 sync → s3://$bucket/"
aws s3 sync $dist "s3://$bucket/" --delete --region $region
if ($LASTEXITCODE -ne 0) { throw "s3 sync falhou" }

Write-Host "==> CloudFront invalidate /* ($distId)"
aws cloudfront create-invalidation `
    --distribution-id $distId `
    --paths "/*" `
    --region $region `
    --query "Invalidation.Id" `
    --output text
if ($LASTEXITCODE -ne 0) { throw "create-invalidation falhou" }

Write-Host "OK: front publicado. Abra frontend_url após a invalidação propagar."
