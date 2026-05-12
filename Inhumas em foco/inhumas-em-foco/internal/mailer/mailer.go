package mailer

import (
	"context"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"inhumas-em-foco/internal/config"
)

type Message struct {
	To      string
	Subject string
	Text    string
}

func Send(ctx context.Context, cfg *config.Config, msg Message) error {
	if cfg == nil || !cfg.SMTPEnabled() {
		return fmt.Errorf("smtp nao configurado")
	}
	to := strings.TrimSpace(msg.To)
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("destinatario invalido: %w", err)
	}
	fromAddr := strings.TrimSpace(cfg.SMTPFrom)
	if _, err := mail.ParseAddress(fromAddr); err != nil {
		return fmt.Errorf("remetente invalido: %w", err)
	}

	host := strings.TrimSpace(cfg.SMTPHost)
	port := strings.TrimSpace(cfg.SMTPPort)
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(host, port)
	var auth smtp.Auth
	if cfg.SMTPUsername != "" || cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, host)
	}

	body := buildMessage(cfg, to, msg)
	errCh := make(chan error, 1)
	go func() {
		errCh <- smtp.SendMail(addr, auth, fromAddr, []string{to}, []byte(body))
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func SendPasswordReset(ctx context.Context, cfg *config.Config, to, resetURL string) error {
	return Send(ctx, cfg, Message{
		To:      to,
		Subject: "Redefinicao de senha - Inhumas em Foco",
		Text: "Voce solicitou a redefinicao de senha do painel Inhumas em Foco.\n\n" +
			"Acesse o link abaixo em ate 30 minutos:\n" + resetURL + "\n\n" +
			"Se voce nao solicitou essa alteracao, ignore esta mensagem.",
	})
}

func buildMessage(cfg *config.Config, to string, msg Message) string {
	from := (&mail.Address{Name: cfg.SMTPFromName, Address: cfg.SMTPFrom}).String()
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + sanitizeHeader(msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.TrimSpace(msg.Text) + "\r\n"
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}
