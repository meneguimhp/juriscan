# Juriscan

Leitura inteligente de publicações juridicas com geração assistida de prazos.

## Estado atual
- Sprint 1 iniciado.
- Backend base implementado em `juriscan-backend/` (auth OTP, RBAC, auditoria e CRM inicial).
- Frontend MVP implementado em `juriscan-frontend/` (login OTP, dashboard comercial, pipeline e edição de leads).
- CI validando backend + frontend em `.github/workflows/ci.yml`.

## Estrutura
- `juriscan-backend/`: API Go do MVP.
- `juriscan-frontend/`: interface React do MVP.
- `docs/product/`: SDD, backlog e roadmap.
- `docs/technical/`: arquitetura, API e plano de validação.
- `infra/`: Terraform da infraestrutura AWS isolada do Juriscan.

## Documentacao principal
- SDD MVP: `docs/product/sdd-mvp.md`
- Backlog técnico 90 dias: `docs/product/backlog-90d.md`
- Kanban inicial 90 dias: `docs/product/kanban-90d.md`
- Roadmap executivo: `docs/product/roadmap.md`
- Arquitetura técnica: `docs/technical/architecture.md`
- Identidade visual (logo): `docs/brand/README.md`

## Rodar backend local
```bash
cd juriscan-backend
go run ./cmd/juriscan
```

## Rodar frontend local
```bash
cd juriscan-frontend
npm install
npm run dev
```

## Usuarios e login (MVP)
- Nao existe auto-cadastro publico.
- Login OTP so funciona para usuario cadastrado e `active`.
- No piloto, a tela avisa quando o e-mail nao esta cadastrado e mostra o codigo de teste quando `LOGIN_TOKEN_ECHO=true`.
- Login ficticio recomendado para demonstracao: `demo@juriscan.local`.
- A primeira tela apos login e a Entrada do Dia: filas de atendimento, leads, publicacoes, prazos e validacoes.
- Menu do piloto: `Entrada do Dia`, `WhatsApp`, `Leads`, `Publicacoes`, `Prazos`, `Conferencia IA` e `Usuarios`.
- Follow-up nao aparece como modulo principal; fica como retorno agendado dentro da rotina de Leads.
- Gestao de usuarios via API admin:
  - `GET /v1/admin/users`
  - `POST /v1/admin/users`
  - `PATCH /v1/admin/users/{id}`
- Variaveis `*_EMAILS` servem apenas para bootstrap inicial.

## Fluxo recomendado para demonstracao
1. Entrar com `demo@juriscan.local`.
2. Abrir `WhatsApp` e registrar uma mensagem recebida; marcar `Gerar lead automaticamente` quando a conversa deve virar oportunidade.
3. Voltar para `Entrada do Dia` para ver a fila comercial e o SLA.
4. Em `Leads`, rodar triagem IA e aplicar/ajustar a sugestao humana.
5. Em `Publicacoes`, cadastrar o texto do despacho/publicacao e executar a analise assistida.
6. Validar o prazo sugerido antes de criar tarefa operacional em `Prazos`.
7. Usar `Conferencia IA` para conferir rastreabilidade das analises de IA.

## WhatsApp (MVP + oficial Meta Cloud)
- Local: `WHATSAPP_PROVIDER=mock`
  - simulacao autenticada via `POST /v1/whatsapp/simulate`
- Oficial: `WHATSAPP_PROVIDER=meta_cloud`
  - exige:
    - `WHATSAPP_META_API_VERSION`
    - `WHATSAPP_META_PHONE_NUMBER_ID`
    - `WHATSAPP_META_ACCESS_TOKEN`
    - `WHATSAPP_META_VERIFY_TOKEN`
  - webhook de verificacao:
    - `GET /v1/whatsapp/webhook`
  - webhook de mensagens:
    - `POST /v1/whatsapp/webhook`
  - envio outbound:
    - `POST /v1/whatsapp/messages/send`

## Checklist de deploy (bloco pronto)
- Infra MVP:
  - tudo versionado e provisionado via Terraform em `infra/`
  - usar VPC/subnet publica existente do SynPlace, sem alterar recursos do SynPlace
  - 1 EC2 exclusiva do Juriscan (`t3.micro` inicialmente; subir para `t3.small` se faltar memoria), Amazon Linux 2023
  - Security Group proprio do Juriscan
  - Elastic IP proprio do Juriscan
  - DNS externo reaproveitando dominio do SynPlace, sem alterar registros existentes
  - criar `juriscan.synplace.com.br` apontando para o Elastic IP do Juriscan
  - Caddy na borda (`80/443`)
  - MySQL 8 local, sem exposicao publica
  - sem backup obrigatorio do MySQL no piloto inicial
  - auto-stop da EC2 via EventBridge Scheduler habilitado (ex.: 22:00 America/Sao_Paulo)
  - auto-start da EC2 via EventBridge Scheduler criado/documentado, mas desativado por padrao
- Backend:
  - `APP_ENV=production`
  - `HTTP_ADDR=127.0.0.1:8080`
  - `DATABASE_DRIVER=mysql`
  - `DATABASE_URL=juriscan:<senha>@tcp(127.0.0.1:3306)/juriscan?parseTime=true`
  - `LOGIN_TOKEN_ECHO=false`
  - `ALLOWED_ORIGINS=https://juriscan.synplace.com.br`
  - `WHATSAPP_PROVIDER=meta_cloud` (se integracao oficial ativa)
  - variaveis Meta Cloud preenchidas
- Frontend:
  - build com `npm run build`
  - publicar `dist/` em `/opt/juriscan/frontend` na EC2
  - servir pelo Caddy no mesmo host
- Validacao minima:
  - `GET /healthz`
  - login OTP de usuario ativo
  - recebimento/simulacao de mensagem WhatsApp
  - envio de mensagem outbound (`/v1/whatsapp/messages/send`)

## Infra AWS atual
- Endpoint publico: `https://juriscan.synplace.com.br`
- DNS: `juriscan.synplace.com.br -> 18.205.92.243`
- EC2: `i-09c3a3ef8076fe5c4` (`t3.micro`)
- VPC/subnet reaproveitadas: `vpc-0c1d54b3667114bcc` / `subnet-0d266f94dd6868b97`
- Artefatos de deploy: `juriscan-prod-artifacts-343150994013`
- Auto-stop habilitado; auto-start criado e desabilitado.

A infraestrutura e a aplicação estão aplicadas. Validações realizadas:
- `GET https://juriscan.synplace.com.br/healthz`
- `POST https://juriscan.synplace.com.br/v1/identity/auth/login`
- `terraform plan -detailed-exitcode` sem mudanças pendentes.

Observação de piloto: `LOGIN_TOKEN_ECHO=true` está ativo para teste sem provedor de e-mail. Desligar antes de uso real com dados sensíveis.
Usuários iniciais do piloto são definidos em `bootstrap_admin_emails` no Terraform; não há auto-cadastro público.


