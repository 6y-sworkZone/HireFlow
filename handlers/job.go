package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hireflow/database"
	"hireflow/middleware"
	"hireflow/models"
	"hireflow/utils"
)

func CreateJob(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := database.DB.Exec(
		`INSERT INTO jobs (title, department, location, salary_min, salary_max, description, requirements, tags, status, workflow_id, creator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.Title, job.Department, job.Location, job.SalaryMin, job.SalaryMax,
		job.Description, job.Requirements, job.Tags, job.Status, job.WorkflowID, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建职位失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetJobList(c *gin.Context) {
	var pag utils.Pagination
	c.ShouldBindQuery(&pag)

	department := c.Query("department")
	status := c.Query("status")
	keyword := c.Query("keyword")

	where := "WHERE 1=1"
	args := []interface{}{}

	if department != "" {
		where += " AND j.department = ?"
		args = append(args, department)
	}
	if status != "" {
		where += " AND j.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (j.title LIKE ? OR j.tags LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM jobs j " + where
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	query := `SELECT j.id, j.title, j.department, j.location, j.salary_min, j.salary_max, 
			  j.tags, j.status, j.created_at, u.real_name as creator_name
			  FROM jobs j LEFT JOIN users u ON j.creator_id = u.id ` + where + ` ORDER BY j.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, pag.PageSize, pag.Offset())

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询职位列表失败")
		return
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var j models.Job
		rows.Scan(&j.ID, &j.Title, &j.Department, &j.Location, &j.SalaryMin, &j.SalaryMax,
			&j.Tags, &j.Status, &j.CreatedAt, &j.CreatorName)
		jobs = append(jobs, j)
	}

	pag.Total = total
	utils.Success(c, utils.PaginatedData{List: jobs, Pagination: pag})
}

func GetJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var job models.Job
	err := database.DB.QueryRow(
		`SELECT j.id, j.title, j.department, j.location, j.salary_min, j.salary_max,
		 j.description, j.requirements, j.tags, j.status, j.workflow_id, j.creator_id,
		 u.real_name as creator_name, j.created_at, j.updated_at
		 FROM jobs j LEFT JOIN users u ON j.creator_id = u.id WHERE j.id = ?`,
		id,
	).Scan(&job.ID, &job.Title, &job.Department, &job.Location, &job.SalaryMin, &job.SalaryMax,
		&job.Description, &job.Requirements, &job.Tags, &job.Status, &job.WorkflowID, &job.CreatorID,
		&job.CreatorName, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "职位不存在")
		return
	}

	utils.Success(c, job)
}

func UpdateJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var job models.Job
	c.ShouldBindJSON(&job)

	_, err := database.DB.Exec(
		`UPDATE jobs SET title = COALESCE(NULLIF(?, ''), title), department = COALESCE(NULLIF(?, ''), department),
		 location = ?, salary_min = ?, salary_max = ?, description = ?, requirements = ?, tags = ?, status = ?, workflow_id = ?, updated_at = ? WHERE id = ?`,
		job.Title, job.Department, job.Location, job.SalaryMin, job.SalaryMax,
		job.Description, job.Requirements, job.Tags, job.Status, job.WorkflowID, time.Now(), id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新职位失败")
		return
	}

	utils.Success(c, nil)
}

func DeleteJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	database.DB.Exec("DELETE FROM jobs WHERE id = ?", id)
	utils.Success(c, nil)
}

func GetJobStats(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var stats models.JobStats
	database.DB.QueryRow(`
		SELECT 
			COUNT(DISTINCT cj.id) as resume_count,
			COUNT(DISTINCT CASE WHEN i.id IS NOT NULL THEN cj.id END) as interview_count,
			COUNT(DISTINCT CASE WHEN o.id IS NOT NULL THEN cj.id END) as offer_count
		FROM candidate_jobs cj
		LEFT JOIN interviews i ON i.candidate_job_id = cj.id
		LEFT JOIN offers o ON o.candidate_job_id = cj.id
		WHERE cj.job_id = ?
	`, id).Scan(&stats.ResumeCount, &stats.InterviewCount, &stats.OfferCount)
	stats.JobID = id

	utils.Success(c, stats)
}

func GetDepartments(c *gin.Context) {
	rows, err := database.DB.Query("SELECT DISTINCT department FROM jobs WHERE department != '' ORDER BY department")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询部门列表失败")
		return
	}
	defer rows.Close()

	var departments []string
	for rows.Next() {
		var d string
		rows.Scan(&d)
		departments = append(departments, d)
	}

	utils.Success(c, departments)
}

func CreateJobTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var tpl models.JobTemplate
	c.ShouldBindJSON(&tpl)

	result, err := database.DB.Exec(
		`INSERT INTO job_templates (name, title, department, location, salary_min, salary_max, description, requirements, tags, creator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tpl.Name, tpl.Title, tpl.Department, tpl.Location, tpl.SalaryMin, tpl.SalaryMax,
		tpl.Description, tpl.Requirements, tpl.Tags, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建模板失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetJobTemplates(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, title, department, location, salary_min, salary_max, description, requirements, tags, created_at FROM job_templates ORDER BY created_at DESC")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询模板列表失败")
		return
	}
	defer rows.Close()

	var templates []models.JobTemplate
	for rows.Next() {
		var t models.JobTemplate
		rows.Scan(&t.ID, &t.Name, &t.Title, &t.Department, &t.Location, &t.SalaryMin, &t.SalaryMax, &t.Description, &t.Requirements, &t.Tags, &t.CreatedAt)
		templates = append(templates, t)
	}

	utils.Success(c, templates)
}

func DeleteJobTemplate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM job_templates WHERE id = ?", id)
	utils.Success(c, nil)
}
