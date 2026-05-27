package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hireflow/config"
	"hireflow/database"
	"hireflow/middleware"
	"hireflow/models"
	"hireflow/utils"
)

func CreateOffer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var offer models.Offer
	c.ShouldBindJSON(&offer)

	if offer.Status == "" {
		offer.Status = "pending_approval"
	}

	result, err := database.DB.Exec(
		`INSERT INTO offers (candidate_job_id, salary, start_date, terms, status, template_id, creator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		offer.CandidateJobID, offer.Salary, offer.StartDate, offer.Terms,
		offer.Status, offer.TemplateID, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建Offer失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetOfferList(c *gin.Context) {
	status := c.Query("status")

	where := "WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		where += " AND o.status = ?"
		args = append(args, status)
	}

	query := `SELECT o.id, o.candidate_job_id, o.salary, o.start_date, o.terms, o.status, o.template_id,
			  o.pdf_path, o.creator_id, o.created_at, o.updated_at,
			  c.name as candidate_name, j.title as job_title
			  FROM offers o
			  LEFT JOIN candidate_jobs cj ON o.candidate_job_id = cj.id
			  LEFT JOIN candidates c ON cj.candidate_id = c.id
			  LEFT JOIN jobs j ON cj.job_id = j.id ` + where + ` ORDER BY o.created_at DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询Offer列表失败")
		return
	}
	defer rows.Close()

	var offers []models.Offer
	for rows.Next() {
		var o models.Offer
		rows.Scan(&o.ID, &o.CandidateJobID, &o.Salary, &o.StartDate, &o.Terms, &o.Status,
			&o.TemplateID, &o.PDFPath, &o.CreatorID, &o.CreatedAt, &o.UpdatedAt,
			&o.CandidateName, &o.JobTitle)
		offers = append(offers, o)
	}

	utils.Success(c, offers)
}

func GetOffer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var o models.Offer
	err := database.DB.QueryRow(
		`SELECT o.id, o.candidate_job_id, o.salary, o.start_date, o.terms, o.status, o.template_id,
		 o.pdf_path, o.creator_id, o.created_at, o.updated_at,
		 c.name as candidate_name, j.title as job_title, c.email, c.phone
		 FROM offers o
		 LEFT JOIN candidate_jobs cj ON o.candidate_job_id = cj.id
		 LEFT JOIN candidates c ON cj.candidate_id = c.id
		 LEFT JOIN jobs j ON cj.job_id = j.id WHERE o.id = ?`,
		id,
	).Scan(&o.ID, &o.CandidateJobID, &o.Salary, &o.StartDate, &o.Terms, &o.Status,
		&o.TemplateID, &o.PDFPath, &o.CreatorID, &o.CreatedAt, &o.UpdatedAt,
		&o.CandidateName, &o.JobTitle, &o.CandidateJobID, &o.Salary)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "Offer不存在")
		return
	}

	utils.Success(c, o)
}

func UpdateOffer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Salary     int    `json:"salary"`
		StartDate  string `json:"start_date"`
		Terms      string `json:"terms"`
		Status     string `json:"status"`
		TemplateID int    `json:"template_id"`
	}
	c.ShouldBindJSON(&req)

	_, err := database.DB.Exec(
		`UPDATE offers SET salary = COALESCE(NULLIF(?, 0), salary), start_date = COALESCE(NULLIF(?, ''), start_date),
		 terms = ?, status = COALESCE(NULLIF(?, ''), status), template_id = ?, updated_at = ? WHERE id = ?`,
		req.Salary, req.StartDate, req.Terms, req.Status, req.TemplateID, time.Now(), id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新Offer失败")
		return
	}

	utils.Success(c, nil)
}

func DeleteOffer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM offer_approval WHERE offer_id = ?", id)
	database.DB.Exec("DELETE FROM offers WHERE id = ?", id)
	utils.Success(c, nil)
}

func CreateOfferApproval(c *gin.Context) {
	offerID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		ApproverIDs []int `json:"approver_ids" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	database.DB.Exec("DELETE FROM offer_approval WHERE offer_id = ?", offerID)

	for i, approverID := range req.ApproverIDs {
		database.DB.Exec(
			"INSERT INTO offer_approval (offer_id, approver_id, \"order\", status) VALUES (?, ?, ?, 'pending')",
			offerID, approverID, i+1,
		)
	}

	database.DB.Exec("UPDATE offers SET status = 'pending_approval', updated_at = ? WHERE id = ?", time.Now(), offerID)

	utils.Success(c, nil)
}

func GetOfferApproval(c *gin.Context) {
	offerID, _ := strconv.Atoi(c.Param("id"))

	rows, err := database.DB.Query(
		`SELECT oa.id, oa.offer_id, oa.approver_id, oa."order", oa.status, oa.comment, oa.approved_at, oa.created_at,
		 u.real_name as approver_name
		 FROM offer_approval oa LEFT JOIN users u ON oa.approver_id = u.id
		 WHERE oa.offer_id = ? ORDER BY oa."order"`,
		offerID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询审批链失败")
		return
	}
	defer rows.Close()

	var approvals []models.OfferApproval
	for rows.Next() {
		var a models.OfferApproval
		rows.Scan(&a.ID, &a.OfferID, &a.ApproverID, &a.Order, &a.Status, &a.Comment,
			&a.ApprovedAt, &a.CreatedAt, &a.ApproverName)
		approvals = append(approvals, a)
	}

	utils.Success(c, approvals)
}

func ApproveOffer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	offerID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&req)

	var order int
	database.DB.QueryRow(
		"SELECT \"order\" FROM offer_approval WHERE offer_id = ? AND approver_id = ? AND status = 'pending' ORDER BY \"order\" LIMIT 1",
		offerID, userID,
	).Scan(&order)

	if order == 0 {
		utils.Error(c, http.StatusBadRequest, "您没有待审批的Offer")
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE offer_approval SET status = 'approved', comment = ?, approved_at = ? WHERE offer_id = ? AND approver_id = ? AND \"order\" = ?",
		req.Comment, now, offerID, userID, order,
	)

	var pendingCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM offer_approval WHERE offer_id = ? AND status = 'pending'",
		offerID,
	).Scan(&pendingCount)

	if pendingCount == 0 {
		database.DB.Exec("UPDATE offers SET status = 'approved', updated_at = ? WHERE id = ?", time.Now(), offerID)
	}

	utils.Success(c, nil)
}

func RejectOffer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	offerID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Comment string `json:"comment" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	database.DB.Exec(
		"UPDATE offer_approval SET status = 'rejected', comment = ?, approved_at = ? WHERE offer_id = ? AND approver_id = ? AND status = 'pending'",
		req.Comment, time.Now(), offerID, userID,
	)

	database.DB.Exec("UPDATE offers SET status = 'rejected', updated_at = ? WHERE id = ?", time.Now(), offerID)

	utils.Success(c, nil)
}

func SendOffer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var offer struct {
		CandidateJobID int
		Salary         int
		StartDate      string
		Terms          string
		TemplateID     int
	}
	database.DB.QueryRow(
		"SELECT candidate_job_id, salary, start_date, terms, template_id FROM offers WHERE id = ?",
		id,
	).Scan(&offer.CandidateJobID, &offer.Salary, &offer.StartDate, &offer.Terms, &offer.TemplateID)

	var candidate struct {
		Name  string
		Email string
	}
	database.DB.QueryRow(
		"SELECT c.name, c.email FROM candidate_jobs cj JOIN candidates c ON cj.candidate_id = c.id WHERE cj.id = ?",
		offer.CandidateJobID,
	).Scan(&candidate.Name, &candidate.Email)

	cfg := config.Load()
	pdfPath := filepath.Join(cfg.UploadDir, "offers", fmt.Sprintf("offer_%d.pdf", id))

	defaultTemplate := `<h1>入职Offer</h1>
<p>尊敬的 {{.Name}}：</p>
<p>经过综合评估，我们很高兴地通知您，您已通过我司的面试，现正式向您发出入职Offer。</p>
<p>职位信息：</p>
<ul>
<li>薪资：{{.Salary}} 元/月</li>
<li>入职日期：{{.StartDate}}</li>
<li>其他条款：{{.Terms}}</li>
</ul>
<p>期待您的加入！</p>`

	data := map[string]interface{}{
		"Name":      candidate.Name,
		"Salary":    offer.Salary,
		"StartDate": offer.StartDate,
		"Terms":     offer.Terms,
	}

	if offer.TemplateID > 0 {
		var htmlTemplate string
		database.DB.QueryRow("SELECT html_template FROM offer_templates WHERE id = ?", offer.TemplateID).Scan(&htmlTemplate)
		if htmlTemplate != "" {
			utils.GenerateOfferPDF(htmlTemplate, data, pdfPath)
		} else {
			utils.GenerateOfferPDF(defaultTemplate, data, pdfPath)
		}
	} else {
		utils.GenerateOfferPDF(defaultTemplate, data, pdfPath)
	}

	database.DB.Exec("UPDATE offers SET status = 'sent', pdf_path = ?, updated_at = ? WHERE id = ?", pdfPath, time.Now(), id)

	utils.Success(c, gin.H{"pdf_path": pdfPath})
}

func UpdateOfferStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	database.DB.Exec("UPDATE offers SET status = ?, updated_at = ? WHERE id = ?", req.Status, time.Now(), id)

	utils.Success(c, nil)
}

func CreateOfferTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var tpl models.OfferTemplate
	c.ShouldBindJSON(&tpl)

	result, err := database.DB.Exec(
		"INSERT INTO offer_templates (name, salary_structure, terms, html_template, creator_id) VALUES (?, ?, ?, ?, ?)",
		tpl.Name, tpl.SalaryStructure, tpl.Terms, tpl.HTMLTemplate, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建模板失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetOfferTemplates(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, salary_structure, terms, html_template, created_at FROM offer_templates ORDER BY created_at DESC")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询模板列表失败")
		return
	}
	defer rows.Close()

	var templates []models.OfferTemplate
	for rows.Next() {
		var t models.OfferTemplate
		rows.Scan(&t.ID, &t.Name, &t.SalaryStructure, &t.Terms, &t.HTMLTemplate, &t.CreatedAt)
		templates = append(templates, t)
	}

	utils.Success(c, templates)
}

func DeleteOfferTemplate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM offer_templates WHERE id = ?", id)
	utils.Success(c, nil)
}

func DownloadOfferPDF(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cfg := config.Load()

	var pdfPath string
	database.DB.QueryRow("SELECT pdf_path FROM offers WHERE id = ?", id).Scan(&pdfPath)

	if pdfPath == "" {
		utils.Error(c, http.StatusNotFound, "PDF文件不存在")
		return
	}

	fullPath := filepath.Join(cfg.UploadDir, "offers", filepath.Base(pdfPath))
	c.File(fullPath)
}

func GetOfferStatuses(c *gin.Context) {
	statuses := []string{"待审批", "已审批", "已发送", "已接受", "已拒绝", "已过期"}
	utils.Success(c, statuses)
}
