# Plano de validação (MVP)

## Objetivo
Validar as dores prioritarias do escritório antes de expandir escopo.

## Bloco 1 - Comercial
1. Cadastrar 30 leads reais/simulados no pipeline.
2. Medir tempo médio para triagem e 1a resposta.
3. Validar que o time não volta para planilha.

## Bloco 2 - IA de triagem
1. Rodar 100 leads historicos.
2. Medir acuracia da classificação inicial por area/urgencia.
3. Medir taxa de override humano.

## Bloco 3 - Publicação e prazo
1. Ingerir publicações por upload/texto.
2. Comparar prazo sugerido x prazo validado pelo humano.
3. Medir percentual tratado no mesmo dia.

## Criterios de sucesso
- Redução de uso de planilha no comercial e operação.
- Ganho de tempo no fluxo de triagem.
- 100% dos prazos com validação humana registrada.

## Validacao automatizada (atual)
### Backend
- `go test ./...`
- Cobertura minima de stores e middleware + helpers HTTP.

### Frontend
- Unit: `npm run test`
- Build: `npm run build`
- E2E (Playwright): `npm run test:e2e`

### Jornadas cobertas por E2E
1. Login OTP + troca de tema + persistencia de sessao.
2. Cadastro e edicao de lead no pipeline.
3. Triagem IA + aplicacao de proximo passo sugerido.
4. Templates + agendamento e conclusao de follow-up.
5. Publicacao -> analise IA -> validacao humana -> criacao de tarefa.
6. Painel de prazos (alteracao de status) + tela de compliance IA.


