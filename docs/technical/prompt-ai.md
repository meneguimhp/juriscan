# Prompt IA (base)

## Triagem de lead
Entrada:
- resumo do contato
- origem
- mensagens iniciais

Saida esperada (JSON):
```json
{
  "area_juridica": "trabalhista|civil|consumidor|outros",
  "urgencia": "baixa|media|alta",
  "potencial_fechamento": 0.0,
  "proximo_passo": "agendar_reuniao|pedir_documentos|descartar|encaminhar_advogado",
  "justificativa": "texto curto"
}
```

## Leitura de publicação
Entrada:
- texto integral da publicação

Saida esperada (JSON):
```json
{
  "tipo_ato": "despacho|decisão|sentenca|intimacao|outro",
  "data_base": "YYYY-MM-DD",
  "prazo_sugerido_dias": 0,
  "responsavel_sugerido": "controller|advogado",
  "nivel_risco": "baixo|medio|alto",
  "justificativa": "texto curto"
}
```

## Regra de segurança
- O sistema sempre exige validação humana para prazo final.
- Prompt e resposta devem ser auditados.


