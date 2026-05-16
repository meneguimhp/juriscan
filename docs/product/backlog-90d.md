# Backlog Técnico 90 Dias - Juriscan

## 1) Objetivo
Transformar o SDD em plano de execução de 90 dias, priorizando as dores do escritório:
1. CRM juridico com WhatsApp para area comercial.
2. IA de triagem de leads.
3. Leitura de publicações com sugestão de prazo.
4. Operação de prazos com controle, dono e cobrança.
5. Redução de risco por uso de IA sem governança.

## 2) Criterio de priorizacao
Priorizacao por impacto no problema real do escritório:
- P0: reduz trabalho manual agora e evita risco de prazo.
- P1: aumenta produtividade e qualidade da operação.
- P2: melhora escala, governança e maturidade.

## 3) Epicos e stories (MVP)

### EP-00 Fundação técnica e segurança (P0)
Objetivo: criar base mínima segura para operar com dados juridicos.

US-001 Auth e perfis base
- Como usuario do escritório, quero login e perfis (admin, controller, advogado, comercial) para acesso controlado.
- Criterios de aceite:
  - Login funcional com sessão.
  - RBAC por modulo.
  - Rotas protegidas por perfil.
  - Somente usuario cadastrado e `active` pode solicitar OTP e concluir login.
  - Sem auto-cadastro publico.

US-001A Gestão administrativa de usuarios
- Como admin, quero cadastrar e gerenciar usuarios do escritorio para controlar acesso sem depender de variavel de ambiente.
- Criterios de aceite:
  - Endpoint admin: `GET/POST/PATCH /v1/admin/users`.
  - Campos: `name`, `email`, `role`, `status`, `last_login_at`.
  - Inativacao bloqueia login imediatamente.
  - `ADMIN_EMAILS` fica apenas como bootstrap inicial.

US-002 Auditoria de eventos críticos
- Como gestor, quero trilha de auditoria para saber quem fez o que.
- Criterios de aceite:
  - Log em criação/edicao de lead, triagem IA, prazo e conclusao.
  - Consulta de auditoria por periodo/usuario.

US-003 Estrutura de deploy e observabilidade base
- Como operação, quero padrão de infra igual SynPlace em conta AWS separada.
- Criterios de aceite:
  - CI/CD com GitHub Actions + OIDC + SSM.
  - Logs centralizados no CloudWatch.
  - Health check de frontend e backend.

### EP-01 CRM de leads juridico (P0)
Objetivo: tirar o comercial da planilha e organizar funil.

US-010 Pipeline Kanban de leads
- Como comercial, quero mover lead entre etapas para controlar avancos.
- Criterios de aceite:
  - Etapas: novo, triado, qualificado, proposta, fechado, perdido.
  - Drag and drop ou acao de mover status.
  - Histórico de mudancas via `lead.history` com eventos `created`, `stage_changed` e `updated`.

US-011 Cadastro mínimo de lead
- Como comercial, quero cadastrar lead rapidamente sem friccao.
- Criterios de aceite:
  - Campos minimos: nome, contato, origem, assunto.
  - Responsável e prioridade.

US-012 SLA de resposta e fila
- Como gestor, quero saber quais leads estao atrasados.
- Criterios de aceite:
  - Indicador de tempo sem resposta.
  - Filtro de "SLA estourado".

### EP-02 WhatsApp integrado ao CRM (P0)
Objetivo: centralizar entrada comercial e vincular atendimento ao lead.

US-020 Caixa de entrada WhatsApp
- Como comercial, quero ver mensagens recebidas em uma fila unica.
- Criterios de aceite:
  - Receber mensagens de WhatsApp no sistema.
  - Marcar conversa como "sem lead" ou "vinculada".

US-021 Vincular conversa a lead
- Como comercial, quero vincular uma conversa ao lead para manter contexto.
- Criterios de aceite:
  - Associacao manual e sugerida por telefone.
  - Histórico de conversa no lead.

US-022 Modelos de resposta e follow-up
- Como comercial, quero usar templates para responder mais rapido.
- Criterios de aceite:
  - Templates por tipo de atendimento.
  - Agendamento de follow-up.

### EP-03 IA de triagem de lead (P0)
Objetivo: reduzir tempo de qualificacao inicial.

US-030 Classificação automatica de lead
- Como controller/comercial, quero classificação por area, urgencia e potencial.
- Criterios de aceite:
  - Saida com categoria, urgencia, score e justificativa.
  - Tempo de processamento < 10s por lead.

US-031 Recomendacao de proximo passo
- Como comercial, quero sugestão acionável para não depender de decisão ad-hoc.
- Criterios de aceite:
  - Sugestoes: solicitar docs, agendar contato, descartar.
  - Botao para aplicar sugestão e registrar resultado.

US-032 Override humano com rastreio
- Como advogado responsável, quero corrigir triagem da IA quando necessário.
- Criterios de aceite:
  - Edicao permitida com motivo.
  - Registro de auditoria da mudanca.

### EP-04 Publicações e interpretacao assistida por IA (P0)
Objetivo: tirar leitura manual isolada e padronizar interpretacao.

US-040 Ingestao de publicação (texto/arquivo)
- Como controller, quero enviar publicação ao sistema para analise.
- Criterios de aceite:
  - Upload de arquivo e entrada de texto.
  - Registro de origem e timestamp.

US-041 Extração assistida de dados juridicos
- Como controller, quero extrair elementos-chave para montar tarefa.
- Criterios de aceite:
  - Campos: tipo de ato, data base, prazo sugerido, risco, responsável sugerido.
  - Exibir confiança da IA.

US-042 Confirmação humana obrigatoria
- Como escritório, quero impedir uso de prazo sem validação humana.
- Criterios de aceite:
  - Status "pendente validação".
  - Apenas usuario autorizado confirma prazo final.

### EP-05 Operação de prazos e tarefas (P0)
Objetivo: transformar publicação em execução operacional controlada.

US-050 Criação automatica de tarefa a partir da publicação
- Como controller, quero gerar tarefa com dono e data limite.
- Criterios de aceite:
  - Tarefa nasce com status inicial, dono e checklist.
  - Relacao direta com publicação.

US-051 Quadro de prazos por responsável
- Como gestor, quero visualizar fila por advogado/controller.
- Criterios de aceite:
  - Filtros por vencimento, responsável e status.
  - Destaque de risco (hoje, amanha, atrasado).

US-052 Alertas e cobrança interna
- Como controller, quero alertar responsável antes do vencimento.
- Criterios de aceite:
  - Alertas configuraveis (D-1, D-0, atraso).
  - Registro de notificacao enviada.

### EP-06 Compliance, LGPD e governança de IA (P1)
Objetivo: reduzir risco juridico e operacional do uso de IA.

US-060 Registro de prompt/resposta por analise
- Como compliance, quero rastrear o que a IA decidiu e por que.
- Criterios de aceite:
  - Persistir prompt, resposta, confiança, modelo e timestamp.
  - Vincular ao item de negocio (lead/publicação).

US-061 Controle de campos sensíveis por perfil
- Como admin, quero limitar acesso a dados críticos.
- Criterios de aceite:
  - Mascaramento/ocultacao por perfil.
  - Auditoria de visualizacao de dados sensíveis.

US-062 Política de retenção
- Como compliance, quero definir prazo de retenção e descarte.
- Criterios de aceite:
  - Regras configuradas por tipo de dado.
  - Processo de descarte logado.

## 4) Sequencia de execução (90 dias)

### Ciclo 1 - Dias 1 a 15 (Fundação P0)
Entrega:
- EP-00 completo (auth, RBAC, auditoria base, pipeline CI/CD).
- Esqueleto do dominio de leads e publicações.

Gate de saida:
- Ambiente dev/hml operacional.
- Trilha de auditoria funcionando em eventos base.

### Ciclo 2 - Dias 16 a 35 (Comercial P0)
Entrega:
- EP-01 completo (CRM lead).
- EP-02 parcial (caixa WhatsApp + vinculo com lead).

Gate de saida:
- Time comercial consegue operar lead sem planilha.

### Ciclo 3 - Dias 36 a 55 (IA comercial P0)
Entrega:
- EP-03 completo (triagem IA + override humano).
- EP-02 completo (templates/follow-up).

Gate de saida:
- Triagem inicial de lead assistida por IA em producao controlada.

### Ciclo 4 - Dias 56 a 75 (Operacional juridico P0)
Entrega:
- EP-04 completo (ingestao e interpretacao de publicações).
- EP-05 parcial (geração de tarefa + quadro de prazos).

Gate de saida:
- Controller para de operar publicação em planilha para casos piloto.

### Ciclo 5 - Dias 76 a 90 (Hardening e piloto)
Entrega:
- EP-05 completo (alertas/cobrança).
- EP-06 mínimo (governança de IA e compliance essencial).
- Ajustes de UX e performance.

Gate de saida:
- Piloto com escritório rodando ponta a ponta:
  - lead -> triagem -> atendimento -> publicação -> prazo -> conclusao.

## 5) Dependencias externas criticas
1. Canal oficial de WhatsApp (API/fornecedor homologado).
2. Definicao de modelo IA e custo por volume.
3. Política interna do escritório para validação de prazo.
4. Conta AWS dedicada (separada do SynPlace) com baseline de segurança.

## 6) Métricas operacionais por ciclo
1. Leads triados em até 15 minutos.
2. Tempo médio de 1a resposta comercial.
3. Publicações processadas no mesmo dia.
4. Prazos validados humanamente sem excecao.
5. Redução percentual de uso de planilha no escritório.

## 7) Backlog posterior (fora dos 90 dias)
1. Captura automatica ampla de publicações por mais fontes.
2. Integracoes com sistemas processuais adicionais.
3. BI juridico e previsao de carga.
4. Recomendacao de estrategia por tipo de caso.

---
Status: backlog técnico pronto para quebrar em tarefas de sprint.

