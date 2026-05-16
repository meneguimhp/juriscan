# API - Juriscan (MVP)

## Base URL local
`http://localhost:8080`

## Regras gerais
- Sessao por cookie `juriscan_session` (HttpOnly) e suporte a `Authorization: Bearer`.
- OTP somente para usuario cadastrado e `active`.
- Sem auto-cadastro publico.
- Perfis: `admin`, `controller`, `lawyer`, `commercial`.

## Persistencia de usuarios
- Usuarios persistidos em SQLite (`DB_PATH`, default `juriscan.db`).
- `ADMIN_EMAILS`, `CONTROLLER_EMAILS`, `LAWYER_EMAILS`, `COMMERCIAL_EMAILS` sao bootstrap inicial.
- Operacao diaria de usuarios deve usar endpoints admin.

## Health

### GET `/healthz`
```json
{
  "status": "ok",
  "env": "development"
}
```

## Autenticacao (OTP)

### POST `/v1/identity/auth/login`
Body:
```json
{ "email": "admin@juriscan.local" }
```

Resposta:
```json
{
  "message": "otp_sent",
  "expires_in_seconds": 600
}
```

Em `development`, com `LOGIN_TOKEN_ECHO=true`, retorna tambem:
```json
{ "token": "123456" }
```

### POST `/v1/identity/auth/verify`
Body:
```json
{
  "email": "admin@juriscan.local",
  "token": "123456"
}
```

Resposta:
```json
{
  "access_token": "....",
  "user": {
    "id": "user-...",
    "email": "admin@juriscan.local",
    "role": "admin"
  }
}
```

### GET `/v1/identity/me`
Retorna usuario da sessao ativa.

### POST `/v1/identity/logout`
Revoga sessao e limpa cookie.

## Admin - Usuarios (`admin`)

### GET `/v1/admin/users`
Lista usuarios.

### POST `/v1/admin/users`
Body:
```json
{
  "name": "Leticia",
  "email": "leticia@juriscan.local",
  "role": "controller",
  "status": "active"
}
```

### PATCH `/v1/admin/users/{id}`
Body parcial:
```json
{
  "role": "lawyer",
  "status": "inactive"
}
```

## CRM de Leads

Permissao de leitura/edicao:
- `admin`, `commercial`, `controller`, `lawyer`

Criacao de lead:
- `admin`, `commercial`

### GET `/v1/crm/leads`
Query opcional:
- `sla=estourado`

Resposta:
```json
{
  "sla_minutes": 30,
  "items": [
    {
      "id": "lead-...",
      "name": "Cliente Teste",
      "phone": "11999999999",
      "stage": "novo",
      "minutes_without_response": 42,
      "sla_status": "estourado",
      "next_step": "Solicitar documentos",
      "next_follow_up_at": "2026-04-14T14:00:00Z",
      "ai_classification": {
        "category": "civil",
        "urgency": "alta",
        "score": 81,
        "justification": "..."
      },
      "history": [
        { "type": "created", "at": "2026-04-14T10:00:00Z" }
      ]
    }
  ]
}
```

### POST `/v1/crm/leads`
Campos minimos:
- `name`
- `phone`

### PATCH `/v1/crm/leads/{id}`
Atualiza dados gerais, incluindo:
- `stage`
- `next_step`

### PATCH `/v1/crm/leads/{id}/stage`
Body:
```json
{ "stage": "qualificado" }
```

### POST `/v1/crm/leads/{id}/triage`
Executa triagem IA heuristica para categoria/urgencia/score/proximo passo.

### PATCH `/v1/crm/leads/{id}/triage/override`
Permissao:
- `admin`, `controller`, `lawyer`

Body:
```json
{
  "reason": "Ajuste humano",
  "category": "trabalhista",
  "urgency": "media",
  "score": 70,
  "justification": "..."
}
```

## Templates e Follow-ups

Permissao:
- `admin`, `commercial`, `controller`, `lawyer`

### GET `/v1/crm/templates`
Lista templates.

### POST `/v1/crm/templates`
Criacao somente:
- `admin`, `commercial`

Body:
```json
{
  "name": "Primeiro contato",
  "channel": "whatsapp",
  "body": "Texto padrao..."
}
```

### GET `/v1/crm/followups`
Query opcional:
- `lead_id=<id>`
- `pending=true`

### POST `/v1/crm/followups`
Body:
```json
{
  "lead_id": "lead-123",
  "template_id": "tpl-123",
  "message": "Texto final",
  "due_at": "2026-04-14T18:00:00Z"
}
```

### PATCH `/v1/crm/followups/{id}`
Body:
```json
{ "status": "concluido" }
```

Status validos:
- `pendente`
- `concluido`
- `cancelado`

## Caixa WhatsApp (MVP + bloco oficial Meta Cloud)

Configuracao por ambiente:
- `WHATSAPP_PROVIDER=mock` (padrao local)
- `WHATSAPP_PROVIDER=meta_cloud` (oficial Meta)

Com `meta_cloud`, configurar:
- `WHATSAPP_META_API_VERSION` (ex.: `v20.0`)
- `WHATSAPP_META_PHONE_NUMBER_ID`
- `WHATSAPP_META_ACCESS_TOKEN`
- `WHATSAPP_META_VERIFY_TOKEN`

### POST `/v1/whatsapp/webhook`
Entrada oficial de mensagem.

Se `WHATSAPP_PROVIDER=mock` e `WHATSAPP_WEBHOOK_TOKEN` estiver configurado:
- Header `X-Webhook-Token: <token>`

Body:
```json
{
  "phone": "11999990000",
  "message": "Oi, preciso de ajuda",
  "contact_name": "Cliente",
  "external_id": "wa-123"
}
```

### GET `/v1/whatsapp/webhook`
Verificacao de webhook da Meta (modo `meta_cloud`).

Query esperada pela Meta:
- `hub.mode`
- `hub.verify_token`
- `hub.challenge`

### POST `/v1/whatsapp/simulate`
Simulacao local autenticada (nao oficial), usada pelo frontend do MVP.

Permissao:
- `admin`, `commercial`

Body:
```json
{
  "phone": "11999990000",
  "message": "Oi, preciso de ajuda",
  "contact_name": "Cliente"
}
```

### GET `/v1/whatsapp/conversations`
Permissao:
- `admin`, `commercial`

Retorna tambem sugestao de vinculo por telefone:
- `suggested_lead_id`
- `suggested_lead_name`

### PATCH `/v1/whatsapp/conversations/{id}`
Permissao:
- `admin`, `commercial`

Body:
```json
{
  "status": "vinculada",
  "lead_id": "lead-abc"
}
```

Status validos:
- `nova`
- `sem_lead`
- `vinculada` (exige `lead_id`)

### POST `/v1/whatsapp/messages/send`
Permissao:
- `admin`, `commercial`

Body:
```json
{
  "to": "11999990000",
  "message": "Mensagem de retorno"
}
```

Com `WHATSAPP_PROVIDER=mock`: retorno simulado.  
Com `WHATSAPP_PROVIDER=meta_cloud`: envio via Graph API.

## Publicacoes e prazos

### GET `/v1/publications`
Permissao:
- `admin`, `controller`, `lawyer`

### POST `/v1/publications`
Permissao:
- `admin`, `controller`, `lawyer`

Body:
```json
{
  "source": "DJE",
  "input_type": "texto",
  "file_name": "",
  "raw_text": "Despacho ..."
}
```

`input_type`:
- `texto`
- `arquivo`

### POST `/v1/publications/{id}/analyze`
Permissao:
- `admin`, `controller`, `lawyer`

Executa extracao assistida (ato, risco, prazo sugerido, dono sugerido, confianca).

### POST `/v1/publications/{id}/validate`
Permissao:
- `admin`, `controller`, `lawyer`

Confirmacao humana obrigatoria e criacao de tarefa de prazo.

Body:
```json
{
  "final_deadline_at": "2026-04-20T18:00:00Z",
  "owner_email": "advogado@juriscan.local",
  "notes": "Prazo validado."
}
```

### GET `/v1/deadlines/tasks`
Permissao:
- `admin`, `controller`, `lawyer`

Query opcional:
- `owner_email=<email>`
- `status=<status>`

### PATCH `/v1/deadlines/tasks/{id}`
Permissao:
- `admin`, `controller`

Body:
```json
{
  "status": "em_execucao",
  "owner_email": "advogado@juriscan.local"
}
```

Status:
- `aberto`
- `em_execucao`
- `concluido`

### GET `/v1/deadlines/alerts`
Permissao:
- `admin`, `controller`, `lawyer`

Alertas dinamicos:
- D-1
- D0
- atrasado

## Compliance IA

### GET `/v1/compliance/ai-logs`
Permissao:
- `admin`, `controller`

Query opcional:
- `feature=<feature>`

Features registradas no MVP:
- `lead_triage`
- `publication_analysis`

### POST `/v1/compliance/retention/run`
Permissao:
- `admin`

Executa retencao:
- remove AI logs expirados
- remove publicacoes fora da janela de retencao

Config:
- `AI_LOG_RETENTION_DAYS`
- `PUBLICATION_RETENTION_DAYS`
