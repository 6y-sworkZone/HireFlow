package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"hireflow/config"
)

func SendEmail(to string, subject string, body string) error {
	cfg := config.Load()

	if cfg.SMTPHost == "" || cfg.SMTPUser == "" {
		return fmt.Errorf("SMTP未配置")
	}

	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}

	mime := "MIME-Version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s",
		from, to, subject, mime, body)

	smtpPort := cfg.SMTPPort
	if smtpPort == "" {
		smtpPort = "587"
	}

	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, smtpPort)

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("SMTP客户端创建失败: %v", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("创建邮件数据失败: %v", err)
	}

	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭邮件数据失败: %v", err)
	}

	return nil
}
