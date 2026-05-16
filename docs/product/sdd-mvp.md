# SDD MVP - Juriscan

## 1) Contexto
Juriscan nasce para resolver duas frentes do escritório:
1. Comercial: captar, qualificar e converter leads com menos trabalho manual.
2. Operacional: tratar publicações e prazos com segurança, rastreabilidade e menos risco.

Contexto real levantado com o escritório:
- Ferramentas atuais de mercado são boas em partes, mas fracas no fluxo completo desejado.
- Controller faz leitura de publicações "na unha" (planilha + repasse manual).
- Quando há dúvida jurídica, uso de IA pública sem controle de segurança/compliance.
- Ferramentas generalistas exigem muito cadastro/manual feed e aumentam carga operacional.

## 2) Dores priorizadas (entrada para o MVP)
1. Perda de tempo na triagem de leads e no atendimento inicial.
2. Falta de CRM juridico enxuto, integrado ao WhatsApp.
3. Alto esforço manual no tratamento de publicações e cobrança de prazo.
4. Risco operacional por dependência de planilha e processo manual.
5. Risco de segurança/compliance por uso de ferramentas de IA sem governança.

## 3) Objetivo do MVP
Entregar um produto que reduza trabalho manual e risco operacional em 90 dias, com foco em:
- Lead-to-intake mais rapido (triagem e qualificacao).
- Publicação-to-prazo com IA assistiva e validação humana.
- Controle de dono, prazo e status por tarefa, com trilha de auditoria.

## 4) Benchmark de mercado (abril/2026)
Referencia de mercado usada para definicao de posicionamento:

### 4.1 Plataformas juridicas completas (fortes em operação processual)
- ADVBOX: proposta de plataforma completa para escritório e camada de IA jurídica.
- Astrea (Aurum): forte em publicações, andamentos e gestão de prazos.
- Projuris ADV: foco em controle de prazos/processos e módulos de captura de intimacoes.
- Legal One (Thomson Reuters): suíte ampla de gestão (contencioso, contratos, governança, financeiro).

### 4.2 Plataformas de atendimento/comercial (fortes em captacao e WhatsApp)
- Chat Juridico (exemplo de categoria): foco em WhatsApp + CRM + triagem de leads com IA.

### 4.3 Leitura critica do benchmark
- Existe oferta forte em "suíte jurídica completa", mas com custo de implantação e operação.
- Existe oferta forte em "CRM + WhatsApp", porem desconectada da rotina de prazos/publicações.
- Oportunidade do Juriscan: juntar as duas pontas em fluxo único e enxuto para escritório médio.
- Tanto termos comerciais quanto FAQ de mercado reforcam que captura/monitoramento não pode ser o único controle de risco; precisa de conferência e processo interno.

## 5) Posicionamento do produto
Juriscan não será um ERP juridico genérico.
Juriscan será um "sistema de operação jurídica enxuta", com 2 motores integrados:
1. Motor Comercial (lead e atendimento)
2. Motor Operacional (publicação, prazo e execução)

Principio de produto:
- IA assistiva com human-in-the-loop obrigatório para atos críticos (prazo e classificação sensível).

## 6) Escopo do MVP

### 6.1 Em escopo (MVP)
1. CRM de Leads (juridico)
- Pipeline Kanban de leads (novo, triado, qualificado, proposta, fechado, perdido).
- Cadastro mínimo e histórico de interações.
- Responsável por lead e SLA de resposta.

2. Integracao WhatsApp (comercial)
- Entrada de conversa em caixa unica do escritório.
- Vinculo da conversa a lead/contato.
- Mensagens modelo e follow-up.

3. IA de triagem de lead
- Classificar area jurídica, urgencia e potencial de fechamento.
- Sugerir proximo passo (agenda, pedido de docs, descarte).
- Registrar justificativa e nível de confiança.

4. Leitura de publicação e interpretacao assistida por IA
- Ingestao inicial por texto/arquivo (MVP inicial sem dependência de captura automatica full).
- Extração de: tipo de ato, responsável sugerido, prazo sugerido, data base, risco.
- Geração de tarefa de prazo para controller/advogado validar.

5. Gestão operacional de prazos e tarefas
- Checklist por tipo de ato.
- Atribuição de responsável, data limite e status.
- Alertas e cobrança interna.
- Trilha de auditoria (quem criou, quem validou, quem concluiu, quando).

### 6.2 Fora do escopo (fase posterior)
- ERP financeiro completo de escritório.
- Automacao integral de peticionamento.
- Integracoes profundas com todos os tribunais no MVP inicial.
- BI juridico avancado.

### 6.3 Experiencia guiada do piloto
Decisao de produto apos avaliacao com o contexto do escritorio: o Juriscan deve parecer uma mesa de trabalho enxuta, nao um ERP juridico.

Entrada padrao apos login:
- A primeira tela sera "Entrada do Dia", com filas objetivas para a controller, comercial e advogado.
- O usuario deve entender o proximo passo sem treinamento longo: mensagem recebida, lead a responder, publicacao a analisar, prazo a validar ou tarefa em aberto.
- O menu deve esconder complexidade de produto e refletir a rotina real:
  - Comercial: WhatsApp e Leads.
  - Operacional: Publicacoes e Prazos.
  - Controle: Conferencia IA e Usuarios apenas para perfis autorizados.
- "Leads" deve permanecer como termo principal porque foi a palavra usada pelo escritorio.
- Follow-up nao deve ser item de menu no piloto; deve aparecer como "retorno" dentro da rotina de Leads.
- O item Leads deve levar o usuario para o pipeline/carteira, enquanto retornos pendentes entram como acao secundaria da Entrada do Dia.

Login e seguranca:
- Nao havera auto-cadastro publico.
- A tela de login deve deixar claro que o acesso exige e-mail previamente cadastrado no piloto.
- O login ficticio padrao para demonstracao sera `demo@juriscan.local`.
- Em ambiente de piloto, o codigo pode ser exibido na tela para teste sem provedor de e-mail; isso deve continuar documentado como excecao temporaria.
- Antes de dados reais sensiveis, `LOGIN_TOKEN_ECHO` deve ser desligado e o envio de OTP por e-mail deve ser configurado.

IA e controle humano:
- A IA sugere triagem, classificacao e prazo, mas nao confirma atos criticos sozinha.
- Publicacao analisada pela IA so vira prazo operacional apos validacao humana.
- A controller pode organizar e cobrar a fila; advogado/admin validam decisoes sensiveis conforme permissao.
- Toda analise relevante de IA deve manter rastreabilidade em auditoria.

## 7) Requisitos funcionais (RF)
RF-01: Cadastrar e mover leads no pipeline.
RF-02: Integrar atendimento de WhatsApp ao CRM.
RF-03: Classificar lead com IA e permitir override humano.
RF-04: Ingerir publicação (texto/arquivo) e extrair prazo sugerido.
RF-05: Criar tarefa de prazo com dono, data, status e alerta.
RF-06: Exigir validação humana para confirmação final de prazo.
RF-07: Registrar log de auditoria em eventos críticos.

## 8) Requisitos não funcionais (RNF)
RNF-01: Segurança e confidencialidade de dados juridicos (LGPD by design).
RNF-02: Disponibilidade alvo MVP: 99.0% mensal.
RNF-03: Tempo de resposta API p95 < 800 ms (sem etapas async de IA).
RNF-04: Observabilidade mínima (logs, métricas, alertas operacionais).
RNF-05: No piloto de menor custo, não há backup obrigatório do MySQL; antes de uso com dados reais sensíveis, definir e testar rotina mínima de backup/restauração.

## 9) Arquitetura alvo AWS de menor custo (MVP enxuto)
Diretriz aprovada:
- Subir o Juriscan sem alterar repositórios, deploys, instâncias, bancos ou serviços do SynPlace.
- Priorizar o menor custo possível para demonstração/piloto.
- Pode reaproveitar a VPC pública já usada pelo SynPlace, desde que os recursos do Juriscan sejam separados por EC2, Security Group, diretórios, serviços e tags próprias.
- Usar uma única máquina pequena na AWS para aplicação e banco no MVP inicial.
- Toda a infraestrutura AWS do Juriscan deve ser declarada via Terraform no próprio repositório do Juriscan, em `infra/`.

### 9.1 Topologia MVP
- VPC/subnet pública existente do SynPlace pode ser reaproveitada para reduzir custo e tempo de provisionamento.
- 1 EC2 pública pequena exclusiva do Juriscan para todo o runtime.
- Sistema operacional: Amazon Linux 2023, conforme Terraform em `infra/`.
- Proxy/TLS: Caddy na própria EC2.
- Frontend: React/Vite buildado como arquivos estáticos e servido pelo Caddy.
- Backend: binário Go rodando como serviço `systemd`, escutando apenas em `127.0.0.1:8080`.
- Banco: MySQL 8 na mesma EC2, escutando apenas em `127.0.0.1:3306`.
- Arquivos/anexos: manter local no MVP; migrar para S3 quando houver volume/necessidade.
- Endereço externo: Elastic IP próprio do Juriscan e DNS/subdomínio apontando para esse IP.
- Para evitar compra de domínio no piloto, reaproveitar a zona/dominio do SynPlace com subdomínio dedicado.
- Sugestão de URL pública: `https://juriscan.synplace.com.br`.
- Não alterar `app.synplace.com.br`, `synplace.com.br`, `www.synplace.com.br` nem registros existentes do SynPlace.
- Criar apenas um novo registro DNS isolado, por exemplo `juriscan.synplace.com.br -> Elastic IP do Juriscan`.

### 9.2 Tamanho inicial sugerido
- EC2: `t3.micro` para menor custo no piloto.
- Evolução imediata: subir para `t3.small` se MySQL, backend ou build disputarem memória.
- Disco: EBS gp3 de 20 GB, criptografado, priorizando custo no piloto.
- Swap: 2 GB para reduzir risco de OOM em build/deploy e MySQL.
- Evolução posterior: subir para `t3.medium` se houver uso real mais intenso.

### 9.3 Rede e segurança
- Security Group:
  - Entrada pública: `80/tcp` e `443/tcp`.
  - SSH: restrito ao IP administrativo ou preferencialmente via AWS SSM Session Manager.
  - MySQL: sem exposição pública.
- Security Group do Juriscan deve ser separado do SynPlace, mesmo usando a mesma VPC/subnet.
- Tags mínimas: `Project=juriscan`, `Environment=prod`, `ManagedBy=terraform`, `CostProfile=low`.
- Caddy termina TLS e encaminha `/api/*` ou `/v1/*` para o backend local.
- Backend não deve ficar exposto diretamente na internet.
- MySQL com usuário próprio da aplicação, sem uso de `root` pelo backend.
- Secrets em arquivo de ambiente protegido (`/etc/juriscan/backend.env`) ou AWS SSM Parameter Store.
- Nenhuma regra nova deve ser aplicada em Security Groups, instâncias ou banco do SynPlace.

### 9.4 Banco de dados
- Banco alvo do MVP em AWS: MySQL.
- Banco local de desenvolvimento pode continuar simples, mas produção deve usar MySQL.
- Variáveis esperadas para produção:
  - `APP_ENV=production`
  - `HTTP_ADDR=127.0.0.1:8080`
  - `DATABASE_DRIVER=mysql`
  - `DATABASE_URL=juriscan:<senha>@tcp(127.0.0.1:3306)/juriscan?parseTime=true`
  - `LOGIN_TOKEN_ECHO=false`
  - `ALLOWED_ORIGINS=https://juriscan.synplace.com.br`
  - variáveis de WhatsApp/IA conforme provedor ativo.
- Observação técnica: o backend atual ainda deve ser validado/adaptado para driver MySQL antes da subida produtiva definitiva, pois o estado inicial do MVP nasceu com persistência local simples.

### 9.5 Deploy
- Infra AWS: `terraform init/plan/apply` em `infra/`.
- Build frontend: `npm ci && npm run build`.
- Publicação frontend: copiar `juriscan-frontend/dist/` para `/opt/juriscan/frontend`.
- Build backend: `go build -o juriscan ./cmd/juriscan`.
- Publicação backend: copiar binário para `/opt/juriscan/backend/juriscan`.
- Serviço `systemd`: `juriscan-backend.service`, usando env file protegido.
- Caddy: servir frontend e fazer reverse proxy para backend local.
- Pipeline pode ser GitHub Actions com SSH/SSM para a EC2, mantendo segredo e permissões isoladas do SynPlace.

### 9.6 Automação liga/desliga da EC2
- Seguir o conceito usado no SynPlace para reduzir custo: EventBridge Scheduler chamando AWS SDK EC2.
- Desligamento automático: previsto para ficar habilitado no piloto.
  - Nome sugerido: `juriscan-stop-app-ec2`.
  - Expressão sugerida: `cron(0 22 * * ? *)`.
  - Timezone: `America/Sao_Paulo`.
  - Ação: `ec2:StopInstances` apenas na EC2 do Juriscan.
- Ligamento automático: criar/documentar a automação, mas deixar desativada por padrão.
  - Nome sugerido: `juriscan-start-app-ec2`.
  - Expressão sugerida: `cron(0 8 * * ? *)`.
  - Timezone: `America/Sao_Paulo`.
  - Estado inicial: `DISABLED`.
  - Ação: `ec2:StartInstances` apenas na EC2 do Juriscan.
- A role do Scheduler deve ter permissão mínima para iniciar/parar somente a instância marcada com tags do Juriscan.
- O deploy manual/CI pode ligar a EC2 antes de publicar nova versão, se ela estiver parada.
- Nenhum agendamento do Juriscan deve incluir instâncias do SynPlace.

### 9.7 Backup e restauração
- Para reduzir custo e complexidade no piloto inicial, não haverá rotina obrigatória de backup do MySQL.
- Dados do piloto devem ser considerados descartáveis até aprovação de produção.
- Antes de uso com dados reais sensíveis, reavaliar backup mínimo (`mysqldump` local/S3) e teste de restauração.

### 9.8 Observabilidade mínima
- Logs do backend via `journalctl` e rotação padrão do systemd.
- Logs do Caddy em arquivo ou journal.
- CloudWatch Agent opcional no MVP, recomendado antes de piloto com uso real.
- Alarmes mínimos:
  - uso de disco > 80%
  - CPU alta sustentada
  - status check da EC2
  - falha do serviço backend.

### 9.9 Fora desta primeira subida
- Multi-AZ.
- RDS separado.
- EC2 privada + NAT instance.
- CloudFront/S3 para frontend.
- Rotina obrigatória de backup MySQL.
- Kubernetes/ECS.
- Integrações profundas com tribunais.

Esses itens ficam como evolução quando o piloto validar volume, valor e necessidade operacional.

### 9.10 Implantação AWS realizada (maio/2026)
Infraestrutura provisionada via Terraform em `infra/`, sem alteração de recursos do SynPlace.

- Conta AWS: `343150994013`.
- Região: `us-east-1`.
- VPC reaproveitada: `vpc-0c1d54b3667114bcc` (`synplace-vpc`).
- Subnet pública reaproveitada: `subnet-0d266f94dd6868b97` (`synplace-public-subnet`).
- EC2 exclusiva Juriscan: `i-09c3a3ef8076fe5c4`.
- Tipo EC2: `t3.micro`.
- IP privado: `10.0.1.147`.
- Elastic IP: `18.205.92.243`.
- Security Group exclusivo: `sg-03d403373dd5beaa4`.
- DNS criado: `juriscan.synplace.com.br -> 18.205.92.243`.
- Auto-stop: `juriscan-prod-stop`, habilitado, `cron(0 22 * * ? *)`, timezone `America/Sao_Paulo`.
- Auto-start: `juriscan-prod-start`, criado desabilitado, `cron(0 8 * * ? *)`, timezone `America/Sao_Paulo`.
- Backend state Terraform: bucket S3 `terraform-juriscan-state`, key `juriscan/infra/terraform.tfstate`.

Observação: esta implantação criou a base de infraestrutura. A publicação produtiva da aplicação, instalação final do MySQL/Caddy e validação do backend com driver MySQL ficam como próximo passo antes de considerar o ambiente pronto para uso real.

### 9.11 Publicação da aplicação realizada (maio/2026)
Aplicação publicada na mesma EC2 por Terraform + SSM, mantendo o isolamento do SynPlace.

- URL pública validada: `https://juriscan.synplace.com.br`.
- Healthcheck público validado: `https://juriscan.synplace.com.br/healthz -> {"env":"production","status":"ok"}`.
- Frontend React/Vite servido pelo Caddy.
- Backend Go rodando como `systemd` em `127.0.0.1:8080`.
- MySQL 8 rodando em Docker, exposto somente em `127.0.0.1:3306`.
- Caddy rodando em Docker com network host e proxy local para `/v1/*` e `/healthz`.
- Bucket privado de artefatos: `juriscan-prod-artifacts-343150994013`.
- Deploy versionado em Terraform: `infra/deploy.tf` e `infra/scripts/deploy-app.ps1`.
- O Terraform liga a EC2 antes do deploy quando ela estiver parada e valida o backend antes de concluir.
- `terraform plan -detailed-exitcode` executado após a publicação: sem mudanças pendentes.

Decisões temporárias do piloto:
- `LOGIN_TOKEN_ECHO=true` para permitir teste sem provedor de e-mail; antes de uso real, desligar e integrar envio de OTP por e-mail.
- Usuários do piloto são criados por bootstrap em `bootstrap_admin_emails` no Terraform, sem auto-cadastro público.
- Login ficticio de demonstracao criado por bootstrap: `demo@juriscan.local`.
- O volume MySQL foi resetado uma vez durante a primeira publicação porque a tentativa anterior havia criado credenciais descartáveis de teste.
- Após a primeira publicação, `reset_mysql_volume_on_deploy=false` deve permanecer desligado para não apagar dados em deploys futuros.

## 10) Segurança, LGPD e governança de IA
1. Dados de cliente nunca devem trafegar em ferramenta pessoal sem controle.
2. Todo uso de IA deve ser registrado com:
- prompt estruturado,
- resposta,
- nível de confiança,
- decisão humana final.
3. Campos sensíveis devem ter controle de acesso por perfil.
4. Retenção e descarte de dados definidos por política.

## 11) KPIs de sucesso do MVP
1. Tempo médio de primeira resposta ao lead.
2. Percentual de leads triados em até 15 min em horário comercial.
3. Percentual de publicações tratadas no mesmo dia.
4. Percentual de prazos com validação humana registrada.
5. Redução do uso de planilhas no fluxo operacional.

## 12) Riscos e mitigacoes
R-01: Erro de interpretacao de prazo pela IA.
- Mitigação: human-in-the-loop obrigatório + checklist de validação.

R-02: Dependência de fonte externa (tribunais/diários/WhatsApp).
- Mitigação: retries, monitoramento e fila de exceções.

R-03: Escopo virar ERP completo cedo demais.
- Mitigação: recorte MVP estrito + backlog priorizado por valor.

## 13) Roadmap sugerido (macro)
Fase 0 (2 semanas):
- Fundação técnica, auth, perfis e trilha de auditoria.

Fase 1 (3-4 semanas):
- CRM lead + WhatsApp + IA de triagem.

Fase 2 (3-4 semanas):
- Ingestao de publicações + IA de prazo + tarefas e alertas.

Fase 3 (2 semanas):
- Piloto controlado com o escritório, ajustes e hardening.

## 14) Fontes de benchmark (mercado)
1. ADVBOX - Software juridico: https://advbox.com.br/software-juridico
2. Astrea (Aurum) - FAQ gestão de prazos: https://astrea.aurum.com.br/pt-BR/articles/2343560-perguntas-frequentes-sobre-a-gestão-de-prazos-no-astrea
3. Aurum (site institucional): https://www.aurum.com.br/
4. Projuris ADV (produto): https://www.projuris.com.br/adv/
5. Projuris - captura de intimacoes: https://store.projuris.com.br/products/captura-de-intimacoes-eletronicas-projuris-adv
6. Projuris - termos de compra: https://store.projuris.com.br/pages/termos-de-compra
7. Legal One (Thomson Reuters): https://www.thomsonreuters.com.br/pt/juridico/legal-one/firm/legal-one-o-software-juridico-essencial-para-seu-escritório.html
8. Chat Juridico (categoria CRM+WhatsApp+IA): https://chatjuridico.com.br/

---
Status: documento base pronto para iniciar detalhamento de arquitetura técnica e backlog de implementacao.

