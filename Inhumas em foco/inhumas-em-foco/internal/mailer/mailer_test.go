package mailer

import (
	"strings"
	"testing"

	"inhumas-em-foco/internal/config"
)

func TestBuildMessageSanitizesSubjectAndIncludesBody(t *testing.T) {
	cfg := &config.Config{
		SMTPFrom:     "contato@example.com",
		SMTPFromName: "Inhumas em Foco",
	}
	body := buildMessage(cfg, "editor@example.com", Message{
		Subject: "Reset\r\nBcc: atacante@example.com",
		Text:    "Link seguro",
	})

	if strings.Contains(body, "\r\nBcc:") || strings.Contains(body, "\nBcc:") {
		t.Fatalf("subject header injection was not removed: %s", body)
	}
	for _, want := range []string{"From: \"Inhumas em Foco\" <contato@example.com>", "To: editor@example.com", "Subject: ResetBcc: atacante@example.com", "Link seguro"} {
		if !strings.Contains(body, want) {
			t.Fatalf("message missing %q: %s", want, body)
		}
	}
}

func TestSendRequiresSMTPConfig(t *testing.T) {
	err := Send(t.Context(), &config.Config{}, Message{To: "editor@example.com", Subject: "Teste", Text: "Texto"})
	if err == nil || !strings.Contains(err.Error(), "smtp nao configurado") {
		t.Fatalf("Send error = %v, want smtp nao configurado", err)
	}
}
