# Arquitetura técnica (MVP)

## Diretriz
Subir o Juriscan com o menor custo possível, sem impacto em repositórios, pipelines, instâncias, bancos ou serviços do SynPlace.
Para o MVP/piloto, a prioridade é simplicidade operacional: reaproveitar a VPC/subnet pública já existente do SynPlace quando conveniente, mas manter os recursos do Juriscan separados.

## Stack
- Frontend: React + Vite.
- Backend: Go (HTTP REST).
- Banco de produção MVP: MySQL 8.
- Runtime MVP: 1 EC2 pública pequena.
- Edge/proxy: Caddy na própria EC2.
- Infraestrutura: Terraform versionado em `infra/` no repo do Juriscan.
- Deploy: GitHub Actions via SSH/SSM para a EC2 isolada.
- Segredos: env file protegido ou AWS SSM Parameter Store.
- Observabilidade: systemd/journal, Caddy logs e alarmes CloudWatch mínimos.
- Backup MySQL: fora do piloto inicial para reduzir custo/complexidade.
- Custo: EventBridge Scheduler para desligar a EC2 automaticamente; ligamento automático desativado por padrão.

## Topologia AWS MVP single-node
- VPC/subnet pública existente do SynPlace pode ser reaproveitada.
- EC2 exclusiva do Juriscan: `t3.micro` inicialmente, Amazon Linux 2023, EBS gp3 criptografado.
- Se faltar memória no piloto, subir para `t3.small`.
- Caddy exposto em `80/443`.
- Backend Go escutando em `127.0.0.1:8080`.
- MySQL escutando em `127.0.0.1:3306`.
- Frontend estático em `/opt/juriscan/frontend`.
- Backend em `/opt/juriscan/backend/juriscan`, gerenciado por `systemd`.
- Security Group próprio do Juriscan, sem exposição pública do MySQL e sem exposição direta do backend.
- Nenhuma regra deve ser adicionada aos Security Groups do SynPlace.
- Tags dos recursos: `Project=juriscan`, `Environment=prod`, `ManagedBy=terraform`, `CostProfile=low`.
- Endereço externo: Elastic IP dedicado ao Juriscan.
- DNS: reaproveitar o domínio do SynPlace com subdomínio dedicado, por exemplo `juriscan.synplace.com.br`.
- Route 53: criar apenas um novo `A record` para `juriscan.synplace.com.br` apontando para o Elastic IP do Juriscan.
- Não alterar registros existentes do SynPlace (`synplace.com.br`, `www.synplace.com.br`, `app.synplace.com.br`).

Fluxo HTTP:
1. Usuário acessa o endereço externo (`https://juriscan.synplace.com.br` no piloto).
2. Caddy serve o frontend.
3. Chamadas `/v1/*` são encaminhadas pelo Caddy para `127.0.0.1:8080`.
4. Backend acessa MySQL local via usuário dedicado.

## Configuração de produção esperada
Variáveis mínimas:
- `APP_ENV=production`
- `HTTP_ADDR=127.0.0.1:8080`
- `DATABASE_DRIVER=mysql`
- `DATABASE_URL=juriscan:<senha>@tcp(127.0.0.1:3306)/juriscan?parseTime=true`
- `LOGIN_TOKEN_ECHO=false`
- `ALLOWED_ORIGINS=https://juriscan.synplace.com.br`
- `WHATSAPP_PROVIDER=mock` no piloto sem Meta Cloud, ou `meta_cloud` com credenciais oficiais.

Observação: o backend atual precisa ser validado/adaptado para MySQL antes da subida produtiva definitiva, pois a persistência inicial nasceu simples para desenvolvimento local.

## Backup no piloto
- Não haverá backup obrigatório do MySQL nesta primeira subida.
- Dados do piloto devem ser tratados como descartáveis.
- Quando houver uso real com dados sensíveis, reabrir decisão de backup antes de colocar operação crítica no sistema.

## Automação de custo da EC2
Mesmo conceito operacional usado no SynPlace:
- `juriscan-stop-app-ec2`
  - Estado inicial: `ENABLED`.
  - Timezone: `America/Sao_Paulo`.
  - Expressão sugerida: `cron(0 22 * * ? *)`.
  - Target: `arn:aws:scheduler:::aws-sdk:ec2:stopInstances`.
  - Payload: lista contendo apenas o Instance ID da EC2 do Juriscan.
- `juriscan-start-app-ec2`
  - Estado inicial: `DISABLED`.
  - Timezone: `America/Sao_Paulo`.
  - Expressão sugerida: `cron(0 8 * * ? *)`.
  - Target: `arn:aws:scheduler:::aws-sdk:ec2:startInstances`.
  - Payload: lista contendo apenas o Instance ID da EC2 do Juriscan.

Regras:
- IAM role do Scheduler com permissão mínima para `ec2:StopInstances` e `ec2:StartInstances` somente na EC2 do Juriscan.
- O agendamento não deve referenciar instâncias, Security Groups ou recursos do SynPlace.
- O deploy manual/CI deve poder ligar a EC2 antes de publicar, caso a instância esteja parada.

## Evolução posterior
Quando o piloto exigir maior disponibilidade ou isolamento:
- separar MySQL em RDS;
- mover frontend para S3/CloudFront;
- separar EC2 pública/privada;
- adicionar NAT, backups gerenciados e alarmes mais completos;
- manter o SynPlace totalmente separado.

## Estado atual (Sprint 1 em andamento)
- Backend base criado com:
  - OTP login
  - sessão via cookie HttpOnly (`juriscan_session`) com fallback Bearer
  - RBAC inicial por perfil
  - auditoria de requests em log JSON
  - CRM de leads inicial em memória
- Frontend MVP criado com login OTP, dashboard comercial, cadastro/edição de lead e pipeline por estágio.
- CI criado para backend e frontend (`go test`, `npm run test`, `npm run build`).

## Estado da infraestrutura AWS
Provisionado via Terraform em `infra/` em maio/2026:
- Conta AWS: `343150994013`.
- Região: `us-east-1`.
- URL planejada: `https://juriscan.synplace.com.br`.
- DNS atual: `juriscan.synplace.com.br -> 18.205.92.243`.
- EC2 exclusiva: `i-09c3a3ef8076fe5c4` (`t3.micro`), IP privado `10.0.1.147`.
- VPC/subnet reaproveitadas sem alteração: `vpc-0c1d54b3667114bcc` / `subnet-0d266f94dd6868b97`.
- Security Group próprio: `sg-03d403373dd5beaa4`.
- Auto-stop habilitado: `juriscan-prod-stop`.
- Auto-start criado e desabilitado: `juriscan-prod-start`.
- Backend state: bucket `terraform-juriscan-state`, key `juriscan/infra/terraform.tfstate`.

Nota operacional: a infraestrutura e a publicação da aplicação estão criadas para avaliação. Antes de uso real com dados sensíveis, desligar echo de OTP, configurar provedor de e-mail, definir backup do MySQL e revisar segredos.

## Estado da publicação AWS
Publicado via Terraform + SSM em maio/2026:
- URL pública: `https://juriscan.synplace.com.br`.
- Healthcheck: `https://juriscan.synplace.com.br/healthz`.
- Frontend: arquivos estáticos Vite em `/opt/juriscan/frontend`, servidos pelo Caddy.
- Backend: `/opt/juriscan/backend/juriscan`, serviço `juriscan-backend.service`, ouvindo em `127.0.0.1:8080`.
- Banco: MySQL 8 em Docker, volume `juriscan-mysql-data`, ouvindo apenas em `127.0.0.1:3306`.
- Proxy: Caddy em Docker com network host, roteando `/v1/*` e `/healthz` para o backend local.
- Bucket privado de artefatos: `juriscan-prod-artifacts-343150994013`.
- Deploy: `infra/deploy.tf` aciona `infra/scripts/deploy-app.ps1`, que compila, empacota, envia ao S3 e executa atualização por SSM.

Nota de piloto: `LOGIN_TOKEN_ECHO=true` está ativo somente para avaliação sem provedor de e-mail. `reset_mysql_volume_on_deploy` deve permanecer `false` depois da primeira subida.

## Decisoes aprovadas
1. WhatsApp: Meta Cloud API direto (custo menor no MVP).
2. IA: uso hibrido por custo/qualidade:
   - triagem de lead: modelo custo baixo
   - publicação/prazo: modelo mais robusto
3. Login: e-mail + OTP em todos os perfis.
4. Escopo MVP: 1 escritório (single tenant).
5. Publicações: início manual (upload/texto), com validação humana obrigatoria.


