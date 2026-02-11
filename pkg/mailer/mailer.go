package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	SendInviteEmail(to, name, inviteToken string) error
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
