package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"hireflow/config"
	"hireflow/database"
	"hireflow/middleware"
	"hireflow/models"
	"hireflow/utils"
)

func CreateCandidate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var candidate models.Candidate
	if err := c.ShouldBindJSON(&candidate); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var exists int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM candidates WHERE (email != '' AND email = ?) OR (phone != '' AND phone = ?)",
		candidate.Email, candidate.Phone,
	).Scan(&exists)
	if exists > 0 {
		utils.Error(c, http.StatusConflict, "候选人已存在（邮箱或电话重复）")
		return
	}

	if candidate.Source == "" {
		candidate.Source = "主动投递"
	}
	if candidate.Status == "" {
		candidate.Status = "pool"
	}

	result, err := database.DB.Exec(
		`INSERT INTO candidates (name, email, phone, current_company, work_years, education, 
		 expected_salary, source, tags, remark, status, creator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		candidate.Name, candidate.Email, candidate.Phone, candidate.CurrentCompany, candidate.WorkYears,
		candidate.Education, candidate.ExpectedSalary, candidate.Source, candidate.Tags, candidate.Remark,
		candidate.Status, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建候选人失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetCandidateList(c *gin.Context) {
	var pag utils.Pagination
	c.ShouldBindQuery(&pag)

	source := c.Query("source")
	keyword := c.Query("keyword")
	status := c.Query("status")

	where := "WHERE 1=1"
	args := []interface{}{}

	if source != "" {
		where += " AND c.source = ?"
		args = append(args, source)
	}
	if status != "" {
		where += " AND c.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (c.name LIKE ? OR c.email LIKE ? OR c.phone LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM candidates c "+where, args...).Scan(&total)

	query := `SELECT c.id, c.name, c.email, c.phone, c.current_company, c.work_years, c.education,
			  c.expected_salary, c.source, c.tags, c.remark, c.status, c.resume_path, c.created_at,
			  u.real_name as creator_name
			  FROM candidates c LEFT JOIN users u ON c.creator_id = u.id ` + where + ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, pag.PageSize, pag.Offset())

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询候选人列表失败")
		return
	}
	defer rows.Close()

	var candidates []models.Candidate
	for rows.Next() {
		var cand models.Candidate
		rows.Scan(&cand.ID, &cand.Name, &cand.Email, &cand.Phone, &cand.CurrentCompany, &cand.WorkYears,
			&cand.Education, &cand.ExpectedSalary, &cand.Source, &cand.Tags, &cand.Remark, &cand.Status,
			&cand.ResumePath, &cand.CreatedAt, &cand.CreatorName)
		candidates = append(candidates, cand)
	}

	pag.Total = total
	utils.Success(c, utils.PaginatedData{List: candidates, Pagination: pag})
}

func GetCandidate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var cand models.Candidate
	err := database.DB.QueryRow(
		`SELECT c.id, c.name, c.email, c.phone, c.current_company, c.work_years, c.education,
		 c.expected_salary, c.source, c.avatar, c.resume_path, c.tags, c.remark, c.status, c.creator_id,
		 u.real_name as creator_name, c.created_at, c.updated_at
		 FROM candidates c LEFT JOIN users u ON c.creator_id = u.id WHERE c.id = ?`,
		id,
	).Scan(&cand.ID, &cand.Name, &cand.Email, &cand.Phone, &cand.CurrentCompany, &cand.WorkYears,
		&cand.Education, &cand.ExpectedSalary, &cand.Source, &cand.Avatar, &cand.ResumePath,
		&cand.Tags, &cand.Remark, &cand.Status, &cand.CreatorID, &cand.CreatorName, &cand.CreatedAt, &cand.UpdatedAt)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "候选人不存在")
		return
	}

	utils.Success(c, cand)
}

func UpdateCandidate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var cand models.Candidate
	c.ShouldBindJSON(&cand)

	_, err := database.DB.Exec(
		`UPDATE candidates SET name = COALESCE(NULLIF(?, ''), name), email = ?, phone = ?,
		 current_company = ?, work_years = ?, education = ?, expected_salary = ?, source = ?,
		 avatar = ?, tags = ?, remark = ?, status = ?, updated_at = ? WHERE id = ?`,
		cand.Name, cand.Email, cand.Phone, cand.CurrentCompany, cand.WorkYears,
		cand.Education, cand.ExpectedSalary, cand.Source, cand.Avatar, cand.Tags,
		cand.Remark, cand.Status, time.Now(), id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新候选人失败")
		return
	}

	utils.Success(c, nil)
}

func DeleteCandidate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM candidates WHERE id = ?", id)
	utils.Success(c, nil)
}

func UploadResume(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cfg := config.Load()

	jobID := c.Query("job_id")
	uploadDir := cfg.UploadDir
	if jobID != "" {
		uploadDir = filepath.Join(cfg.UploadDir, "job_"+jobID)
	}
	os.MkdirAll(uploadDir, 0755)

	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "获取文件失败")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		utils.Error(c, http.StatusBadRequest, "仅支持 PDF 和 Word 格式")
		return
	}

	filename := strconv.Itoa(id) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	relPath := filepath.Join(strings.TrimPrefix(uploadDir, "."), filename)
	database.DB.Exec("UPDATE candidates SET resume_path = ?, updated_at = ? WHERE id = ?", relPath, time.Now(), id)

	utils.Success(c, gin.H{"path": relPath})
}

func DownloadResume(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cfg := config.Load()

	var resumePath string
	database.DB.QueryRow("SELECT resume_path FROM candidates WHERE id = ?", id).Scan(&resumePath)

	if resumePath == "" {
		utils.Error(c, http.StatusNotFound, "简历不存在")
		return
	}

	fullPath := cfg.UploadDir + strings.TrimPrefix(resumePath, "/uploads")
	c.File(fullPath)
}

func GetCandidatePool(c *gin.Context) {
	var pag utils.Pagination
	c.ShouldBindQuery(&pag)

	keyword := c.Query("keyword")

	where := "WHERE c.status = 'pool'"
	args := []interface{}{}

	if keyword != "" {
		where += " AND (c.name LIKE ? OR c.email LIKE ? OR c.phone LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM candidates c "+where, args...).Scan(&total)

	query := `SELECT c.id, c.name, c.email, c.phone, c.current_company, c.work_years, c.education,
			  c.expected_salary, c.source, c.tags, c.created_at
			  FROM candidates c ` + where + ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, pag.PageSize, pag.Offset())

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询候选人池失败")
		return
	}
	defer rows.Close()

	var candidates []models.Candidate
	for rows.Next() {
		var cand models.Candidate
		rows.Scan(&cand.ID, &cand.Name, &cand.Email, &cand.Phone, &cand.CurrentCompany,
			&cand.WorkYears, &cand.Education, &cand.ExpectedSalary, &cand.Source, &cand.Tags, &cand.CreatedAt)
		candidates = append(candidates, cand)
	}

	pag.Total = total
	utils.Success(c, utils.PaginatedData{List: candidates, Pagination: pag})
}

func GetSources(c *gin.Context) {
	sources := []string{"招聘网站", "内推", "猎头", "主动投递"}
	utils.Success(c, sources)
}

func CheckDuplicate(c *gin.Context) {
	email := c.Query("email")
	phone := c.Query("phone")

	var count int
	query := "SELECT COUNT(*) FROM candidates WHERE 1=1"
	args := []interface{}{}

	if email != "" {
		query += " AND email = ?"
		args = append(args, email)
	}
	if phone != "" {
		query += " AND phone = ?"
		args = append(args, phone)
	}

	database.DB.QueryRow(query, args...).Scan(&count)
	utils.Success(c, gin.H{"duplicate": count > 0, "count": count})
}
