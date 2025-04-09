package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/wneessen/go-mail"

	ht "html/template"
	tt "text/template"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	client     *mail.Client
	sender     string
	retryCount int
	retryDelay time.Duration
}

func New(cfg data.Config) (*Mailer, error) {
	mailOpts := []mail.Option{
		mail.WithPort(cfg.MailPort),
		mail.WithTimeout(5 * time.Second),
	}
	if !cfg.APIEnv.IsDev() {
		mailOpts = append(
			mailOpts,
			mail.WithSMTPAuth(mail.SMTPAuthLogin),
			mail.WithUsername(cfg.MailUser),
			mail.WithPassword(cfg.MailPswd),
			mail.WithTLSPolicy(mail.TLSMandatory),
		)
	} else {
		mailOpts = append(
			mailOpts,
			mail.WithTLSPolicy(mail.NoTLS),
		)
	}
	client, err := mail.NewClient(
		cfg.MailHost,
		mailOpts...,
	)
	if err != nil {
		return nil, err
	}

	mailer := &Mailer{
		client:     client,
		sender:     cfg.MailSndr,
		retryCount: cfg.MailSendRetries,
		retryDelay: cfg.MailRetryDelay,
	}

	return mailer, nil
}

func (m *Mailer) Send(args ...any) error {
	slog.Debug("Sending email", "args", args)
	recipient, ok := args[0].(types.Email)
	if !ok {
		return fmt.Errorf("invalid recipient type at args index 0")
	}

	templateFile, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("invalid templateFile type at args index 0")
	}

	data := args[2]

	textTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlTmpl, err := ht.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()

	err = msg.To(string(recipient))
	if err != nil {
		return err
	}

	err = msg.From(m.sender)
	if err != nil {
		return err
	}

	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	for i := range m.retryCount {
		err = m.client.DialAndSend(msg)
		if err == nil {
			break
		}
		if i != m.retryCount {
			time.Sleep(m.retryDelay)
		}
	}

	if err != nil {
		slog.Error("Failed to send email", "error", err)
		return err
	} else {
		slog.Debug("Sent registration email", "email", recipient)
	}
	return nil
}
