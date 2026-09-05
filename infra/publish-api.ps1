# Publish API → ECR (linux/arm64) + force deploy ECS.
# Pré: terraform apply; Docker Desktop / buildx; AWS CLI autenticado.
# Uso (a partir de infra/): .\publish-api.ps1
# Ou: .\publish-api.ps1 -SkipDeploy

param(
    [switch]$SkipDeploy
)

$ErrorActionPreference = "Stop"
$InfraDir = $PSScriptRoot
$RepoRoot = Split-Path -Parent $InfraDir
$BackendDir = Join-Path $RepoRoot "backend"

Push-Location $InfraDir
try {
    $region = terraform output -raw aws_region
    $ecrUrl = terraform output -raw ecr_repository_url
    $cluster = terraform output -raw ecs_cluster_name
    $service = terraform output -raw ecs_service_name
}
finally {
    Pop-Location
}

$registry = ($ecrUrl -split "/")[0]
$image = "${ecrUrl}:latest"

Write-Host "==> ECR login ($registry)"
aws ecr get-login-password --region $region |
    docker login --username AWS --password-stdin $registry
if ($LASTEXITCODE -ne 0) { throw "docker login falhou" }

Write-Host "==> Build linux/arm64 → $image"
docker buildx build --platform linux/arm64 -t $image --push $BackendDir
if ($LASTEXITCODE -ne 0) { throw "docker buildx falhou" }

if ($SkipDeploy) {
    Write-Host "==> Skip deploy (-SkipDeploy)"
    exit 0
}

Write-Host "==> Force new deployment ECS ($cluster / $service)"
aws ecs update-service `
    --cluster $cluster `
    --service $service `
    --force-new-deployment `
    --region $region `
    --query "service.serviceName" `
    --output text
if ($LASTEXITCODE -ne 0) { throw "ecs update-service falhou" }

Write-Host "OK: imagem publicada; deploy forçado. Aguarde health em api_url."
