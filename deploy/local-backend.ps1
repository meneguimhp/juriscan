param(
    [string]$AppEnv = "development",
    [string]$HttpAddr = ":8080",
    [string]$WhatsAppProvider = "mock",
    [string]$WhatsAppWebhookToken = "",
    [string]$WhatsAppMetaApiVersion = "v20.0",
    [string]$WhatsAppMetaPhoneNumberId = "",
    [string]$WhatsAppMetaAccessToken = "",
    [string]$WhatsAppMetaVerifyToken = "",
    [string]$AdminEmails = "admin@juriscan.local",
    [string]$ControllerEmails = "leticia@juriscan.local"
)

$ErrorActionPreference = "Stop"

$env:APP_ENV = $AppEnv
$env:HTTP_ADDR = $HttpAddr
$env:LOGIN_TOKEN_ECHO = "true"
$env:WHATSAPP_PROVIDER = $WhatsAppProvider
$env:WHATSAPP_WEBHOOK_TOKEN = $WhatsAppWebhookToken
$env:WHATSAPP_META_API_VERSION = $WhatsAppMetaApiVersion
$env:WHATSAPP_META_PHONE_NUMBER_ID = $WhatsAppMetaPhoneNumberId
$env:WHATSAPP_META_ACCESS_TOKEN = $WhatsAppMetaAccessToken
$env:WHATSAPP_META_VERIFY_TOKEN = $WhatsAppMetaVerifyToken
$env:ADMIN_EMAILS = $AdminEmails
$env:CONTROLLER_EMAILS = $ControllerEmails

Set-Location (Join-Path $PSScriptRoot "..\juriscan-backend")
go run ./cmd/juriscan
