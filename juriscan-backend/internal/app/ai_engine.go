package app

import (
	"fmt"
	"strings"
	"time"
)

const aiDefaultModel = "gpt-4.1-mini-simulado"

func classifyLeadHeuristic(lead Lead) LeadAIClassification {
	subject := strings.ToLower(strings.TrimSpace(lead.Subject))
	origin := strings.ToLower(strings.TrimSpace(lead.Origin))

	category := "consultivo"
	urgency := "media"
	score := 55
	suggestedAction := "agendar contato"
	justificationParts := []string{"classificacao baseada no assunto e origem"}

	highUrgencyHints := []string{"urgente", "liminar", "prisao", "bloqueio", "demissao", "audiencia"}
	for _, hint := range highUrgencyHints {
		if strings.Contains(subject, hint) {
			urgency = "alta"
			score = 82
			suggestedAction = "priorizar analise imediata"
			justificationParts = append(justificationParts, "assunto contem termo critico")
			break
		}
	}

	if strings.Contains(subject, "trabalh") {
		category = "trabalhista"
		score += 6
	}
	if strings.Contains(subject, "contrat") || strings.Contains(subject, "civil") {
		category = "civel"
		score += 4
	}
	if strings.Contains(subject, "penal") || strings.Contains(subject, "criminal") {
		category = "penal"
		score += 8
	}
	if origin == "whatsapp" {
		score += 3
		justificationParts = append(justificationParts, "origem whatsapp com contato ativo")
	}
	if score > 95 {
		score = 95
	}
	if score < 20 {
		score = 20
	}
	confidence := 0.62
	if urgency == "alta" {
		confidence = 0.78
	}

	return LeadAIClassification{
		Category:        category,
		Urgency:         urgency,
		Score:           score,
		Justification:   strings.Join(justificationParts, "; "),
		SuggestedAction: suggestedAction,
		Confidence:      confidence,
		Model:           aiDefaultModel,
		GeneratedAt:     time.Now().UTC(),
		// prompt/response sao registrados no AILogStore.
		OverrideReason:  "",
		OverriddenBy:    "",
		OverriddenAt:    nil,
	}
}

func buildLeadTriagePromptAndResponse(lead Lead, classification LeadAIClassification) (string, string) {
	prompt := fmt.Sprintf(
		"Triar lead juridico. nome=%s; assunto=%s; origem=%s; etapa=%s",
		strings.TrimSpace(lead.Name),
		strings.TrimSpace(lead.Subject),
		strings.TrimSpace(lead.Origin),
		strings.TrimSpace(lead.Stage),
	)
	response := fmt.Sprintf(
		"categoria=%s; urgencia=%s; score=%d; sugestao=%s; justificativa=%s",
		classification.Category,
		classification.Urgency,
		classification.Score,
		classification.SuggestedAction,
		classification.Justification,
	)
	return prompt, response
}

func analyzePublicationHeuristic(publication Publication) PublicationAnalysis {
	text := strings.ToLower(strings.TrimSpace(publication.RawText))
	now := time.Now().UTC()

	actType := "despacho"
	risk := "medio"
	days := 5
	owner := publication.CreatedBy
	if owner == "" {
		owner = "controller@juriscan.local"
	}

	if strings.Contains(text, "sentenca") {
		actType = "sentenca"
		days = 15
	}
	if strings.Contains(text, "intimacao") {
		actType = "intimacao"
		days = 5
	}
	if strings.Contains(text, "embargos") || strings.Contains(text, "recurso") {
		days = 10
		risk = "alto"
	}
	if strings.Contains(text, "48 horas") || strings.Contains(text, "urgente") || strings.Contains(text, "liminar") {
		days = 2
		risk = "alto"
	}

	confidence := 0.66
	if risk == "alto" {
		confidence = 0.8
	}

	prompt := fmt.Sprintf("Extrair dados juridicos de publicacao: %s", truncateForPrompt(publication.RawText, 240))
	response := fmt.Sprintf("ato=%s; prazo_dias=%d; risco=%s; dono=%s", actType, days, risk, owner)

	return PublicationAnalysis{
		ActType:               actType,
		BaseDate:              now,
		SuggestedDeadlineDays: days,
		SuggestedDeadlineAt:   now.AddDate(0, 0, days),
		Risk:                  risk,
		SuggestedOwnerEmail:   owner,
		Confidence:            confidence,
		Model:                 aiDefaultModel,
		Prompt:                prompt,
		Response:              response,
		GeneratedAt:           now,
	}
}

func truncateForPrompt(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
