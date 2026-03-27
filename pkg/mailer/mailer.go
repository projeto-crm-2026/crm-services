package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mailer interface {
	SendInviteEmail(to, name, inviteToken string) error
	SendPaymentSuccessEmail(to, name, planName string, amountCents int, currency string) error
	SendPaymentFailedEmail(to, name, planName, reason string) error
	SendUsageWarningEmail(to, name, resource string, current, limit int, pct int) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string // exemplo: "https://crmnosso.com"
}

type smtpMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) Mailer {
	return &smtpMailer{cfg: cfg}
}

func sanitizeHeader(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("invalid SMTP header value")
	}
	return nil
}

func (m *smtpMailer) send(to, subject, body string) error {
	if err := sanitizeHeader(to); err != nil {
		return err
	}
	if err := sanitizeHeader(subject); err != nil {
		return err
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	// gambiarrazinha pra rodar local sem auth
	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}

func (m *smtpMailer) SendInviteEmail(to, name, inviteToken string) error {
	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", m.cfg.BaseURL, inviteToken)

	// depois vamos ver qual vai ser o body do email e o subject
	subject := "Você foi convidado a participar de uma organização!"
	body := fmt.Sprintf(`Olá %s,

Você foi convidado a participar da plataforma CRM.

Clique no link abaixo para definir sua senha e ativar sua conta:

%s

Este link expirará em 72 horas.

Se você não esperava este convite, por favor, ignore este e-mail.
`, name, inviteLink)

	return m.send(to, subject, body)
}

func (m *smtpMailer) SendPaymentSuccessEmail(to, name, planName string, amountCents int, currency string) error {
	amount := fmt.Sprintf("%.2f", float64(amountCents)/100)

	subject := "Pagamento confirmado - CRM Services"
	body := fmt.Sprintf(`Olá %s,

Seu pagamento foi confirmado com sucesso!

Plano: %s
Valor: %s %s

Obrigado por utilizar o CRM Services.
`, name, planName, currency, amount)

	return m.send(to, subject, body)
}

func (m *smtpMailer) SendPaymentFailedEmail(to, name, planName, reason string) error {
	subject := "Falha no pagamento - CRM Services"
	body := fmt.Sprintf(`Olá %s,

Houve uma falha no processamento do pagamento da sua assinatura.

Plano: %s
Motivo: %s

Por favor, verifique seus dados de pagamento para evitar a interrupção do serviço.

Acesse: %s/settings/billing
`, name, planName, reason, m.cfg.BaseURL)

	return m.send(to, subject, body)
}

func (m *smtpMailer) SendUsageWarningEmail(to, name, resource string, current, limit int, pct int) error {
	subject := fmt.Sprintf("Alerta de uso: %s a %d%% do limite - CRM Services", resource, pct)
	body := fmt.Sprintf(`Olá %s,

Sua organização está utilizando %d%% do limite de %s.

Uso atual: %d / %d

Considere fazer upgrade do seu plano para evitar bloqueios.

Acesse: %s/settings/billing
`, name, pct, resource, current, limit, m.cfg.BaseURL)

	return m.send(to, subject, body)
}
