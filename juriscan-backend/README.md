# Juriscan Backend

API inicial do MVP (Sprint 1):
- Login por OTP de e-mail
- Sessão Bearer token
- RBAC basico por perfil
- Trilha de auditoria em log JSON
- CRUD inicial de leads (memoria)
- Fila comercial com SLA de primeira resposta
- Caixa WhatsApp MVP (webhook + classificacao)
- Gestao de usuarios por admin (sem auto-cadastro)

## Rodar local
1. Copiar `.env.example` para `.env` (opcional).
2. Exportar variaveis no terminal.
3. Executar:

```bash
go run ./cmd/juriscan
```

## Endpoints base
- `GET /healthz`
- `POST /v1/identity/auth/login`
- `POST /v1/identity/auth/verify`
- `GET /v1/identity/me` (Bearer)
- `GET /v1/crm/leads` (Bearer, role admin/commercial)
- `POST /v1/crm/leads` (Bearer, role admin/commercial)
- `PATCH /v1/crm/leads/{id}` (Bearer, role admin/commercial)
- `PATCH /v1/crm/leads/{id}/stage` (Bearer, role admin/commercial)
- `GET /v1/whatsapp/webhook` (verificacao Meta Cloud API)
- `POST /v1/whatsapp/webhook` (entrada oficial de mensagens)
- `POST /v1/whatsapp/simulate` (simulacao local de entrada)
- `GET /v1/whatsapp/conversations` (Bearer, role admin/commercial)
- `PATCH /v1/whatsapp/conversations/{id}` (Bearer, role admin/commercial)
- `POST /v1/whatsapp/messages/send` (Bearer, role admin/commercial)
- `GET /v1/admin/users` (Bearer, role admin)
- `POST /v1/admin/users` (Bearer, role admin)
- `PATCH /v1/admin/users/{id}` (Bearer, role admin)

## Notas
- Em `development` com `LOGIN_TOKEN_ECHO=true`, o OTP volta na resposta de login.
- Em producao, manter `LOGIN_TOKEN_ECHO=false`.
- `DB_PATH` define o arquivo SQLite (usuarios persistidos em banco).
- `ALLOWED_ORIGINS` define as origens CORS permitidas, separadas por virgula.
- `LEAD_SLA_MINUTES` controla o tempo maximo (minutos) para primeira resposta de lead.
- `WHATSAPP_PROVIDER=mock` mantem o modo simulado (MVP local).
- `WHATSAPP_PROVIDER=meta_cloud` habilita webhook/verificacao oficial da Meta.
- Com `meta_cloud`, configure:
  - `WHATSAPP_META_API_VERSION` (ex.: `v20.0`)
  - `WHATSAPP_META_PHONE_NUMBER_ID`
  - `WHATSAPP_META_ACCESS_TOKEN`
  - `WHATSAPP_META_VERIFY_TOKEN`
- Se `WHATSAPP_PROVIDER=mock` e `WHATSAPP_WEBHOOK_TOKEN` estiver preenchido, enviar `X-Webhook-Token`.
- Apenas usuario cadastrado e ativo no banco pode logar (OTP).
- `ADMIN_EMAILS`, `CONTROLLER_EMAILS`, `LAWYER_EMAILS` e `COMMERCIAL_EMAILS` servem apenas como bootstrap inicial.


