package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"hireflow/config"
	"hireflow/database"
	"hireflow/models"
)

func GetSMTPConfig() models.SMTPConfig {
	var cfg models.SMTPConfig

	rows, err := database.DB.Query("SELECT key, value FROM settings WHERE key LIKE 'smtp_%'")
	if err != nil {
		envCfg := config.Load()
		return models.SMTPConfig{
			Host: envCfg.SMTPHost,
			Port: envCfg.SMTPPort,
			User: envCfg.SMTPUser,
			Pass: envCfg.SMTPPass,
			From: envCfg.SMTPFrom,
		}
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		rows.Scan(&key, &value)
		settings[key] = value
	}

	cfg.Host = settings["smtp_host"]
	cfg.Port = settings["smtp_port"]
	cfg.User = settings["smtp_user"]
	cfg.Pass = settings["smtp_pass"]
	cfg.From = settings["smtp_from"]
	cfg.Security = settings["smtp_security"]

	if cfg.Host == "" {
		envCfg := config.Load()
		cfg.Host = envCfg.SMTPHost
		cfg.Port = envCfg.SMTPPort
		cfg.User = envCfg.SMTPUser
		cfg.Pass = envCfg.SMTPPass
		cfg.From = envCfg.SMTPFrom
	}

	return cfg
}

func SendEmail(to string, subject string, body string) error {
	cfg := GetSMTPConfig()
	return SendEmailWithConfig(cfg, to, subject, body)
}

func SendEmailWithConfig(cfg models.SMTPConfig, to string, subject string, body string) error {
	if cfg.Host == "" || cfg.User == "" {
		return fmt.Errorf("SMTP未配置")
	}

	from := cfg.From
	if from == "" {
		from = cfg.User
	}

	mime := "MIME-Version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s",
		from, to, subject, mime, body)

	smtpPort := cfg.Port
	if smtpPort == "" {
		smtpPort = "587"
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, smtpPort)

	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
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
