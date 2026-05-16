$ErrorActionPreference = "Stop"
$env:PYTHONUTF8 = "1"

function Require-Env {
    param([string]$Name)
    $value = [Environment]::GetEnvironmentVariable($Name)
    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "Missing required environment variable: $Name"
    }
    return $value
}

function Invoke-Checked {
    param(
        [string]$FilePath,
        [string[]]$ArgumentList
    )
    & $FilePath @ArgumentList
    if ($LASTEXITCODE -ne 0) {
        throw "$FilePath failed with exit code $LASTEXITCODE"
    }
}

$region = Require-Env "AWS_REGION"
$instanceId = Require-Env "INSTANCE_ID"
$bucket = Require-Env "ARTIFACT_BUCKET"
$fqdn = Require-Env "APP_FQDN"
$mysqlAppPassword = Require-Env "MYSQL_APP_PASSWORD"
$mysqlRootPassword = Require-Env "MYSQL_ROOT_PASSWORD"
$adminEmails = Require-Env "ADMIN_EMAILS"
$loginTokenEcho = Require-Env "LOGIN_TOKEN_ECHO"
$whatsAppProvider = Require-Env "WHATSAPP_PROVIDER"
$resetMysqlVolume = Require-Env "RESET_MYSQL_VOLUME"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$infraDir = Split-Path -Parent $scriptDir
$repoRoot = Split-Path -Parent $infraDir
$buildRoot = Join-Path $infraDir ".deploy"
$staging = Join-Path $buildRoot "staging"
$packagePath = Join-Path $buildRoot "juriscan-deploy.tar.gz"
$remoteScriptPath = Join-Path $buildRoot "remote-deploy.sh"
$parametersPath = Join-Path $buildRoot "ssm-parameters.json"

if (Test-Path $buildRoot) {
    Remove-Item -LiteralPath $buildRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $staging | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $staging "backend") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $staging "frontend") | Out-Null

Invoke-Checked "aws" @("ec2", "start-instances", "--region", $region, "--instance-ids", $instanceId, "--output", "json")
Invoke-Checked "aws" @("ec2", "wait", "instance-running", "--region", $region, "--instance-ids", $instanceId)

$ssmOnline = $false
for ($i = 0; $i -lt 40; $i++) {
    $pingStatus = aws ssm describe-instance-information --region $region --filters "Key=InstanceIds,Values=$instanceId" --query "InstanceInformationList[0].PingStatus" --output text
    if ($LASTEXITCODE -eq 0 -and $pingStatus -eq "Online") {
        $ssmOnline = $true
        break
    }
    Start-Sleep -Seconds 6
}
if (-not $ssmOnline) {
    throw "SSM did not become online for instance $instanceId"
}

Push-Location (Join-Path $repoRoot "juriscan-backend")
try {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Invoke-Checked "go" @("build", "-o", (Join-Path $staging "backend\juriscan"), "./cmd/juriscan")
}
finally {
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}

Push-Location (Join-Path $repoRoot "juriscan-frontend")
try {
    Invoke-Checked "npm.cmd" @("install")
    $env:VITE_API_BASE_URL = "https://$fqdn"
    Invoke-Checked "npm.cmd" @("run", "build")
    Copy-Item -Path ".\dist\*" -Destination (Join-Path $staging "frontend") -Recurse -Force
}
finally {
    Remove-Item Env:\VITE_API_BASE_URL -ErrorAction SilentlyContinue
    Pop-Location
}

Invoke-Checked "tar" @("-czf", $packagePath, "-C", $staging, ".")

$remoteScript = @'
#!/usr/bin/env bash
set -euo pipefail

: "${APP_FQDN:?APP_FQDN required}"
: "${MYSQL_APP_PASSWORD:?MYSQL_APP_PASSWORD required}"
: "${MYSQL_ROOT_PASSWORD:?MYSQL_ROOT_PASSWORD required}"
: "${ADMIN_EMAILS:?ADMIN_EMAILS required}"
: "${LOGIN_TOKEN_ECHO:?LOGIN_TOKEN_ECHO required}"
: "${WHATSAPP_PROVIDER:?WHATSAPP_PROVIDER required}"
: "${RESET_MYSQL_VOLUME:?RESET_MYSQL_VOLUME required}"

dnf update -y
dnf install -y docker awscli rsync tar gzip
systemctl enable docker
systemctl start docker

mkdir -p /opt/juriscan/backend /opt/juriscan/frontend /opt/juriscan/caddy /etc/juriscan
chmod 750 /etc/juriscan

rm -rf /tmp/juriscan-deploy
mkdir -p /tmp/juriscan-deploy
tar -xzf /tmp/juriscan-deploy.tar.gz -C /tmp/juriscan-deploy

install -m 0755 /tmp/juriscan-deploy/backend/juriscan /opt/juriscan/backend/juriscan
rsync -a --delete /tmp/juriscan-deploy/frontend/ /opt/juriscan/frontend/

if [ "$RESET_MYSQL_VOLUME" = "true" ]; then
  docker rm -f juriscan-mysql >/dev/null 2>&1 || true
  docker volume rm juriscan-mysql-data >/dev/null 2>&1 || true
fi

if docker ps -a --format '{{.Names}}' | grep -qx 'juriscan-mysql'; then
  docker start juriscan-mysql >/dev/null
else
  docker run -d \
    --name juriscan-mysql \
    --restart unless-stopped \
    -e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
    -e MYSQL_DATABASE=juriscan \
    -e MYSQL_USER=juriscan \
    -e MYSQL_PASSWORD="$MYSQL_APP_PASSWORD" \
    -p 127.0.0.1:3306:3306 \
    -v juriscan-mysql-data:/var/lib/mysql \
    mysql:8.4
fi

mysql_ready=false
for i in {1..80}; do
  if docker exec juriscan-mysql mysqladmin ping -h 127.0.0.1 -uroot -p"$MYSQL_ROOT_PASSWORD" --silent >/dev/null 2>&1; then
    mysql_ready=true
    break
  fi
  sleep 3
done
if [ "$mysql_ready" != "true" ]; then
  docker logs --tail 80 juriscan-mysql || true
  exit 20
fi

cat >/etc/juriscan/backend.env <<EOF
APP_ENV=production
HTTP_ADDR=127.0.0.1:8080
DATABASE_DRIVER=mysql
DATABASE_URL=juriscan:${MYSQL_APP_PASSWORD}@tcp(127.0.0.1:3306)/juriscan?parseTime=true
ALLOWED_ORIGINS=https://${APP_FQDN}
LOGIN_TOKEN_ECHO=${LOGIN_TOKEN_ECHO}
ADMIN_EMAILS=${ADMIN_EMAILS}
WHATSAPP_PROVIDER=${WHATSAPP_PROVIDER}
EOF
chmod 640 /etc/juriscan/backend.env

cat >/etc/systemd/system/juriscan-backend.service <<'EOF'
[Unit]
Description=Juriscan backend
After=docker.service network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=/etc/juriscan/backend.env
WorkingDirectory=/opt/juriscan/backend
ExecStart=/opt/juriscan/backend/juriscan
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable juriscan-backend
systemctl restart juriscan-backend

backend_ready=false
for i in {1..40}; do
  if curl -fsS http://127.0.0.1:8080/healthz >/tmp/juriscan-healthz.json 2>/tmp/juriscan-healthz.err; then
    backend_ready=true
    break
  fi
  sleep 3
done
if [ "$backend_ready" != "true" ]; then
  journalctl -u juriscan-backend -n 120 --no-pager || true
  cat /tmp/juriscan-healthz.err || true
  exit 21
fi

cat >/opt/juriscan/caddy/Caddyfile <<EOF
${APP_FQDN} {
  encode zstd gzip

  route {
    reverse_proxy /v1/* 127.0.0.1:8080
    reverse_proxy /healthz 127.0.0.1:8080

    root * /srv
    try_files {path} /index.html
    file_server
  }
}
EOF

if docker ps -a --format '{{.Names}}' | grep -qx 'juriscan-caddy'; then
  docker rm -f juriscan-caddy >/dev/null
fi

docker run -d \
  --name juriscan-caddy \
  --restart unless-stopped \
  --network host \
  -v /opt/juriscan/caddy/Caddyfile:/etc/caddy/Caddyfile:ro \
  -v /opt/juriscan/frontend:/srv:ro \
  -v juriscan-caddy-data:/data \
  -v juriscan-caddy-config:/config \
  caddy:2

echo "backend=$(cat /tmp/juriscan-healthz.json)"
systemctl is-active juriscan-backend
docker ps --filter name=juriscan
'@

Set-Content -LiteralPath $remoteScriptPath -Value $remoteScript -Encoding ascii

$artifactKey = "deploy/juriscan-deploy.tar.gz"
$scriptKey = "deploy/remote-deploy.sh"
Invoke-Checked "aws" @("s3", "cp", $packagePath, "s3://$bucket/$artifactKey", "--region", $region)
Invoke-Checked "aws" @("s3", "cp", $remoteScriptPath, "s3://$bucket/$scriptKey", "--region", $region)

$commands = @(
    "aws s3 cp s3://$bucket/$artifactKey /tmp/juriscan-deploy.tar.gz --region $region",
    "aws s3 cp s3://$bucket/$scriptKey /tmp/juriscan-remote-deploy.sh --region $region",
    "chmod +x /tmp/juriscan-remote-deploy.sh",
    "APP_FQDN='$fqdn' MYSQL_APP_PASSWORD='$mysqlAppPassword' MYSQL_ROOT_PASSWORD='$mysqlRootPassword' ADMIN_EMAILS='$adminEmails' LOGIN_TOKEN_ECHO='$loginTokenEcho' WHATSAPP_PROVIDER='$whatsAppProvider' RESET_MYSQL_VOLUME='$resetMysqlVolume' /tmp/juriscan-remote-deploy.sh"
)

@{ commands = $commands } | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $parametersPath -Encoding ascii

$sendPath = Join-Path $buildRoot "ssm-send.json"
aws ssm send-command --region $region --instance-ids $instanceId --document-name "AWS-RunShellScript" --comment "Deploy Juriscan app via Terraform" --parameters "file://$parametersPath" --output json > $sendPath
if ($LASTEXITCODE -ne 0) {
    throw "aws ssm send-command failed with exit code $LASTEXITCODE"
}

$send = Get-Content $sendPath -Raw | ConvertFrom-Json
$commandId = $send.Command.CommandId
aws ssm wait command-executed --region $region --command-id $commandId --instance-id $instanceId
$waitExit = $LASTEXITCODE

$invocationPath = Join-Path $buildRoot "ssm-invocation.json"
aws ssm get-command-invocation --region $region --command-id $commandId --instance-id $instanceId --output json > $invocationPath
if ($LASTEXITCODE -ne 0) {
    throw "aws ssm get-command-invocation failed with exit code $LASTEXITCODE"
}

$invocation = Get-Content $invocationPath -Raw | ConvertFrom-Json
Write-Host $invocation.StandardOutputContent
if ($waitExit -ne 0 -or $invocation.Status -ne "Success") {
    Write-Host $invocation.StandardErrorContent
    throw "SSM deploy failed with status $($invocation.Status)"
}
