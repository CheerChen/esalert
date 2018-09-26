package actions

import (
	"net/smtp"

	"github.com/domodwyer/mailyak"
	"go.uber.org/zap"
	
	"github.com/CheerChen/esalert/logger"
)

// 邮件动作
type Mail struct {
	To      []string `mapstructure:"to" json:"to"`
	Subject string   `mapstructure:"subject" json:"subject"`
	Content string   `mapstructure:"content" json:"content"`
}

// 发送邮件
func (w *Mail) Do() error {
	auth := smtp.PlainAuth(
		"",
		conf.Action.MailUsername,
		conf.Action.MailPwd,
		conf.Action.MailHost,
	)
	mail := mailyak.New(conf.Action.MailHost+":25", auth)
	mail.To(w.To...)
	mail.From(conf.Action.MailUsername)
	mail.Subject(w.Subject)
	mail.HTML().Set(w.Content)

	logger.Info("mail sending request", zap.String("content", w.Content))

	if err := mail.Send(); err != nil {
		return err
	}
	return nil
}
