package app

import "strings"

func maskLeadByRole(item Lead, role string) Lead {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "commercial":
		return item
	default:
		item.Phone = maskPhone(item.Phone)
		item.OwnerEmail = maskEmail(item.OwnerEmail)
		return item
	}
}

func maskConversationByRole(item WhatsAppConversation, role string) WhatsAppConversation {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "commercial":
		return item
	default:
		item.Phone = maskPhone(item.Phone)
		item.ContactName = maskName(item.ContactName)
		return item
	}
}

func maskPublicationByRole(item Publication, role string) Publication {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "controller":
		return item
	default:
		item.RawText = "conteudo sensivel ocultado para este perfil"
		return item
	}
}

func maskPhone(value string) string {
	value = normalizePhone(value)
	if len(value) < 4 {
		return value
	}
	return "******" + value[len(value)-4:]
}

func maskEmail(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return value
	}
	local := parts[0]
	if len(local) <= 2 {
		return "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}

func maskName(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 1 {
		return value
	}
	return value[:1] + "***"
}
