package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hireflow/database"
	"hireflow/middleware"
	"hireflow/models"
	"hireflow/utils"
)

func GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	rows, err := database.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		rows.Scan(&key, &value)
		settings[key] = value
	}

	smtpSettings := map[string]string{
		"smtp_host":     settings["smtp_host"],
		"smtp_port":     settings["smtp_port"],
		"smtp_user":     settings["smtp_user"],
		"smtp_pass":     settings["smtp_pass"],
		"smtp_from":     settings["smtp_from"],
		"smtp_security": settings["smtp_security"],
	}

	utils.Success(c, smtpSettings)
}

func SaveSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	settings := map[string]string{
		"smtp_host":     req["smtp_host"],
		"smtp_port":     req["smtp_port"],
		"smtp_user":     req["smtp_user"],
		"smtp_pass":     req["smtp_pass"],
		"smtp_from":     req["smtp_from"],
		"smtp_security": req["smtp_security"],
	}

	for key, value := range settings {
		_, err := database.DB.Exec(`
			INSERT INTO settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
		`, key, value, value)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "保存失败")
			return
		}
	}

	utils.Success(c, nil)
}

func TestSMTPConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req struct {
		Host      string `json:"smtp_host"`
		Port      string `json:"smtp_port"`
		User      string `json:"smtp_user"`
		Pass      string `json:"smtp_pass"`
		From      string `json:"smtp_from"`
		Security  string `json:"smtp_security"`
		TestEmail string `json:"test_email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.TestEmail == "" {
		utils.Error(c, http.StatusBadRequest, "请输入测试邮件地址")
		return
	}

	cfg := models.SMTPConfig{
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Pass:     req.Pass,
		From:     req.From,
		Security: req.Security,
	}

	subject := "HireFlow 邮件测试"
	body := `
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>🎉 邮件测试成功！</h2>
			<p>恭喜您，HireFlow 招聘管理系统的 SMTP 配置已正确设置。</p>
			<p>此邮件是一封测试邮件，用于验证 SMTP 配置是否正常工作。</p>
			<hr style="margin: 20px 0;">
			<p style="color: #888; font-size: 12px;">此邮件由 HireFlow 招聘管理系统自动发送</p>
		</div>`

	if err := utils.SendEmailWithConfig(cfg, req.TestEmail, subject, body); err != nil {
		utils.Error(c, http.StatusInternalServerError, "发送失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}
