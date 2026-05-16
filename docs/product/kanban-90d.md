# Kanban Inicial (90 dias) - Juriscan

## Como usar
- `Status`: To Do, Doing, Done.
- `Prioridade`: P0 (critico), P1 (importante), P2 (evolutivo).
- `Estimativa`: pontos (1, 2, 3, 5, 8).

## To Do
| ID | Item | Epico | Prioridade | Estimativa | Dono sugerido | Dependencias |
|---|---|---|---|---:|---|---|
| US-003 | Pipeline deploy + observabilidade base | EP-00 | P0 | 5 | DevOps | - |
| US-070 | Integracao oficial Meta Cloud API (sem simulador) | EP-02 | P0 | 8 | Backend/DevOps | US-020 |
| US-071 | Templates aprovados + envio outbound via WhatsApp | EP-02 | P0 | 5 | Backend/Produto | US-070 |
| US-072 | Dashboard operacional de SLA por dono (com filtros salvos) | EP-05 | P1 | 5 | Frontend | US-051 |
| US-073 | Alertas ativos por e-mail/WhatsApp interno | EP-05 | P1 | 5 | Backend | US-052 |

## Doing
| ID | Item | Epico | Prioridade | Estimativa | Dono sugerido |
|---|---|---|---|---:|---|
| US-074 | Hardening de UX e internacionalizacao PT-BR fina | EP-00 | P1 | 3 | Frontend |
| US-075 | Documentacao operacional de suporte e runbook de incidentes | EP-00 | P1 | 3 | Produto/DevOps |

## Done
| ID | Item | Epico | Prioridade | Estimativa | Evidencia |
|---|---|---|---|---:|---|
| US-001 | Auth e perfis base | EP-00 | P0 | 5 | `internal/identity/auth` + `/v1/identity/auth/*` |
| US-001A | Gestao administrativa de usuarios | EP-00 | P0 | 3 | `/v1/admin/users` + pagina `Usuarios` |
| US-002 | Auditoria de eventos criticos | EP-00 | P0 | 3 | middleware + eventos `lead/*`, `publication/*`, `deadline/*` |
| US-010 | Pipeline Kanban de leads | EP-01 | P0 | 5 | `PATCH /v1/crm/leads/{id}/stage` + historico append-only |
| US-011 | Cadastro minimo de lead | EP-01 | P0 | 3 | `POST /v1/crm/leads` + formulario React |
| US-012 | SLA de resposta e fila | EP-01 | P0 | 3 | `GET /v1/crm/leads?sla=estourado` + painel SLA |
| US-020 | Caixa de entrada WhatsApp | EP-02 | P0 | 8 | `/v1/whatsapp/webhook`, `/v1/whatsapp/conversations` |
| US-021 | Vincular conversa a lead | EP-02 | P0 | 3 | `PATCH /v1/whatsapp/conversations/{id}` + sugestao por telefone |
| US-022 | Templates e follow-up | EP-02 | P1 | 3 | `/v1/crm/templates`, `/v1/crm/followups` + tela dedicada |
| US-030 | Classificacao automatica de lead (IA) | EP-03 | P0 | 5 | `POST /v1/crm/leads/{id}/triage` |
| US-031 | Proximo passo sugerido por IA | EP-03 | P0 | 3 | campo `next_step` + acao "Aplicar sugestao" |
| US-032 | Override humano com rastreio | EP-03 | P0 | 2 | `PATCH /v1/crm/leads/{id}/triage/override` + auditoria |
| US-040 | Ingestao de publicacao (texto/arquivo) | EP-04 | P0 | 5 | `POST /v1/publications` |
| US-041 | Extracao assistida (prazo e risco) | EP-04 | P0 | 8 | `POST /v1/publications/{id}/analyze` |
| US-042 | Confirmacao humana obrigatoria | EP-04 | P0 | 3 | `POST /v1/publications/{id}/validate` |
| US-050 | Gerar tarefa de prazo automaticamente | EP-05 | P0 | 5 | validacao cria tarefa em `/v1/deadlines/tasks` |
| US-051 | Quadro de prazos por responsavel | EP-05 | P0 | 3 | pagina `Prazos e Alertas` |
| US-052 | Alertas e cobranca interna | EP-05 | P0 | 5 | `GET /v1/deadlines/alerts` |
| US-060 | Registro prompt/resposta IA | EP-06 | P1 | 3 | `GET /v1/compliance/ai-logs` |
| US-061 | Controle de campos sensiveis por perfil | EP-06 | P1 | 3 | mascaramento por role em lead/conversation/publication |
| US-062 | Politica de retencao | EP-06 | P1 | 3 | `POST /v1/compliance/retention/run` + envs de retencao |

## Sequencia sugerida por sprint
### Sprint 1 (dias 1-15)
- US-001, US-001A, US-002, US-011.

### Sprint 2 (dias 16-30)
- US-010, US-012, US-020.

### Sprint 3 (dias 31-45)
- US-021, US-022, US-030, US-031, US-032.

### Sprint 4 (dias 46-60)
- US-040, US-041, US-042.

### Sprint 5 (dias 61-75)
- US-050, US-051, US-052.

### Sprint 6 (dias 76-90)
- US-060, US-061, US-062 + hardening.

## Regra de foco
- Nenhum item P1 inicia enquanto houver P0 bloqueado das dores do escritorio.
