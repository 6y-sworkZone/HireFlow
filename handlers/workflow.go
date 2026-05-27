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

var DefaultStages = `[{"id":1,"name":"简历筛选"},{"id":2,"name":"电话面试"},{"id":3,"name":"技术面试"},{"id":4,"name":"HR面试"},{"id":5,"name":"Offer"},{"id":6,"name":"入职"}]`

func CreateWorkflow(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var wf models.Workflow
	c.ShouldBindJSON(&wf)

	if wf.Stages == "" {
		wf.Stages = DefaultStages
	}

	result, err := database.DB.Exec(
		"INSERT INTO workflows (name, description, stages, creator_id) VALUES (?, ?, ?, ?)",
		wf.Name, wf.Description, wf.Stages, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建流程失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetWorkflowList(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, description, stages, created_at FROM workflows ORDER BY id")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询流程列表失败")
		return
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var wf models.Workflow
		rows.Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Stages, &wf.CreatedAt)
		workflows = append(workflows, wf)
	}

	if len(workflows) == 0 {
		workflows = append(workflows, models.Workflow{
			ID:          0,
			Name:        "默认流程",
			Description: "系统默认招聘流程",
			Stages:      DefaultStages,
		})
	}

	utils.Success(c, workflows)
}

func GetWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if id == 0 {
		utils.Success(c, models.Workflow{ID: 0, Name: "默认流程", Stages: DefaultStages})
		return
	}

	var wf models.Workflow
	err := database.DB.QueryRow(
		"SELECT id, name, description, stages, creator_id, created_at FROM workflows WHERE id = ?",
		id,
	).Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Stages, &wf.CreatorID, &wf.CreatedAt)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "流程不存在")
		return
	}

	utils.Success(c, wf)
}

func UpdateWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var wf models.Workflow
	c.ShouldBindJSON(&wf)

	_, err := database.DB.Exec(
		"UPDATE workflows SET name = COALESCE(NULLIF(?, ''), name), description = ?, stages = ? WHERE id = ?",
		wf.Name, wf.Description, wf.Stages, id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新流程失败")
		return
	}

	utils.Success(c, nil)
}

func DeleteWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM workflows WHERE id = ?", id)
	utils.Success(c, nil)
}

func AssignCandidateToJob(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		CandidateID int `json:"candidate_id" binding:"required"`
		JobID       int `json:"job_id" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	var exists int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM candidate_jobs WHERE candidate_id = ? AND job_id = ?",
		req.CandidateID, req.JobID,
	).Scan(&exists)
	if exists > 0 {
		utils.Error(c, http.StatusConflict, "候选人已关联到此职位")
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO candidate_jobs (candidate_id, job_id, current_stage, status, assigned_to) VALUES (?, ?, 1, 'active', ?)",
		req.CandidateID, req.JobID, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "关联候选人失败")
		return
	}

	cjID, _ := result.LastInsertId()

	database.DB.Exec("UPDATE candidates SET status = 'active' WHERE id = ?", req.CandidateID)

	database.DB.Exec(
		"INSERT INTO stage_history (candidate_job_id, from_stage, to_stage, operator_id, evaluation) VALUES (?, 0, 1, ?, '进入招聘流程')",
		cjID, userID,
	)

	utils.Success(c, gin.H{"id": cjID})
}

func GetCandidateJobs(c *gin.Context) {
	jobID := c.Query("job_id")
	candidateID := c.Query("candidate_id")
	status := c.Query("status")

	where := "WHERE 1=1"
	args := []interface{}{}

	if jobID != "" {
		where += " AND cj.job_id = ?"
		args = append(args, jobID)
	}
	if candidateID != "" {
		where += " AND cj.candidate_id = ?"
		args = append(args, candidateID)
	}
	if status != "" {
		where += " AND cj.status = ?"
		args = append(args, status)
	}

	query := `SELECT cj.id, cj.candidate_id, cj.job_id, cj.current_stage, cj.status, cj.assigned_to,
			  cj.created_at, cj.updated_at, c.name as candidate_name, c.email, c.phone, c.avatar,
			  j.title as job_title, j.department
			  FROM candidate_jobs cj
			  LEFT JOIN candidates c ON cj.candidate_id = c.id
			  LEFT JOIN jobs j ON cj.job_id = j.id ` + where + ` ORDER BY cj.updated_at DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询关联列表失败")
		return
	}
	defer rows.Close()

	type CandidateJobItem struct {
		models.CandidateJob
		Email      string `json:"email"`
		Phone      string `json:"phone"`
		Avatar     string `json:"avatar"`
		Department string `json:"department"`
	}

	var list []CandidateJobItem
	for rows.Next() {
		var item CandidateJobItem
		rows.Scan(&item.ID, &item.CandidateID, &item.JobID, &item.CurrentStage, &item.Status,
			&item.AssignedTo, &item.CreatedAt, &item.UpdatedAt, &item.CandidateName, &item.Email,
			&item.Phone, &item.Avatar, &item.JobTitle, &item.Department)
		list = append(list, item)
	}

	utils.Success(c, list)
}

func GetCandidateJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var cj models.CandidateJob
	err := database.DB.QueryRow(
		`SELECT cj.id, cj.candidate_id, cj.job_id, cj.current_stage, cj.status, cj.assigned_to,
		 cj.created_at, cj.updated_at, c.name as candidate_name, j.title as job_title
		 FROM candidate_jobs cj
		 LEFT JOIN candidates c ON cj.candidate_id = c.id
		 LEFT JOIN jobs j ON cj.job_id = j.id WHERE cj.id = ?`,
		id,
	).Scan(&cj.ID, &cj.CandidateID, &cj.JobID, &cj.CurrentStage, &cj.Status, &cj.AssignedTo,
		&cj.CreatedAt, &cj.UpdatedAt, &cj.CandidateName, &cj.JobTitle)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "关联记录不存在")
		return
	}

	utils.Success(c, cj)
}

func MoveStage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cjID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		ToStage    int    `json:"to_stage" binding:"required"`
		Evaluation string `json:"evaluation"`
	}
	c.ShouldBindJSON(&req)

	var currentStage int
	database.DB.QueryRow("SELECT current_stage FROM candidate_jobs WHERE id = ?", cjID).Scan(&currentStage)

	_, err := database.DB.Exec(
		"UPDATE candidate_jobs SET current_stage = ?, updated_at = ? WHERE id = ?",
		req.ToStage, time.Now(), cjID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新阶段失败")
		return
	}

	database.DB.Exec(
		"INSERT INTO stage_history (candidate_job_id, from_stage, to_stage, operator_id, evaluation) VALUES (?, ?, ?, ?, ?)",
		cjID, currentStage, req.ToStage, userID, req.Evaluation,
	)

	utils.Success(c, nil)
}

func GetStageHistory(c *gin.Context) {
	cjID := c.Query("candidate_job_id")

	rows, err := database.DB.Query(
		`SELECT sh.id, sh.candidate_job_id, sh.from_stage, sh.to_stage, sh.operator_id, sh.evaluation, sh.created_at,
		 u.real_name as operator_name
		 FROM stage_history sh LEFT JOIN users u ON sh.operator_id = u.id
		 WHERE sh.candidate_job_id = ? ORDER BY sh.created_at`,
		cjID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询阶段历史失败")
		return
	}
	defer rows.Close()

	var history []models.StageHistory
	for rows.Next() {
		var h models.StageHistory
		rows.Scan(&h.ID, &h.CandidateJobID, &h.FromStage, &h.ToStage, &h.OperatorID,
			&h.Evaluation, &h.CreatedAt, &h.OperatorName)
		history = append(history, h)
	}

	utils.Success(c, history)
}

func GetKanbanView(c *gin.Context) {
	jobID := c.Query("job_id")
	if jobID == "" {
		utils.Error(c, http.StatusBadRequest, "请指定职位ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT cj.id, cj.candidate_id, cj.job_id, cj.current_stage, cj.status,
		 c.name as candidate_name, c.email, c.phone, c.avatar, c.tags,
		 c.expected_salary, c.current_company, c.work_years
		 FROM candidate_jobs cj
		 LEFT JOIN candidates c ON cj.candidate_id = c.id
		 WHERE cj.job_id = ? AND cj.status = 'active'
		 ORDER BY cj.current_stage, cj.updated_at DESC`,
		jobID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询看板数据失败")
		return
	}
	defer rows.Close()

	type KanbanCard struct {
		CandidateJobID int    `json:"candidate_job_id"`
		CandidateID    int    `json:"candidate_id"`
		JobID          int    `json:"job_id"`
		Status         string `json:"status"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		Avatar         string `json:"avatar"`
		Tags           string `json:"tags"`
		ExpectedSalary int    `json:"expected_salary"`
		CurrentCompany string `json:"current_company"`
		WorkYears      int    `json:"work_years"`
		CurrentStage   int    `json:"current_stage"`
	}

	var cards []KanbanCard
	for rows.Next() {
		var card KanbanCard
		rows.Scan(&card.CandidateJobID, &card.CandidateID, &card.JobID, &card.CurrentStage,
			&card.Status, &card.Name, &card.Email, &card.Phone, &card.Avatar, &card.Tags,
			&card.ExpectedSalary, &card.CurrentCompany, &card.WorkYears)
		cards = append(cards, card)
	}

	kanban := map[int][]KanbanCard{}
	for _, card := range cards {
		kanban[card.CurrentStage] = append(kanban[card.CurrentStage], card)
	}

	utils.Success(c, kanban)
}

func RejectCandidate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cjID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Reason   string `json:"reason" binding:"required"`
		SendMail bool   `json:"send_mail"`
	}
	c.ShouldBindJSON(&req)

	_, err := database.DB.Exec(
		"UPDATE candidate_jobs SET status = 'rejected', updated_at = ? WHERE id = ?",
		time.Now(), cjID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "拒绝候选人失败")
		return
	}

	var currentStage int
	database.DB.QueryRow("SELECT current_stage FROM candidate_jobs WHERE id = ?", cjID).Scan(&currentStage)

	database.DB.Exec(
		"INSERT INTO stage_history (candidate_job_id, from_stage, to_stage, operator_id, evaluation) VALUES (?, ?, 0, ?, ?)",
		cjID, currentStage, userID, "拒绝: "+req.Reason,
	)

	utils.Success(c, nil)
}

func GetRejectionReasons(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, reason, template FROM rejection_reasons ORDER BY id")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询拒绝原因失败")
		return
	}
	defer rows.Close()

	var reasons []models.RejectionReason
	for rows.Next() {
		var r models.RejectionReason
		rows.Scan(&r.ID, &r.Reason, &r.Template)
		reasons = append(reasons, r)
	}

	if len(reasons) == 0 {
		defaultReasons := []string{"经验不符合", "薪资不匹配", "技术能力不足", "文化不匹配", "候选人放弃"}
		for _, reason := range defaultReasons {
			reasons = append(reasons, models.RejectionReason{Reason: reason, Template: "很遗憾，经过综合评估，我们决定暂不推进您的面试流程。感谢您对我们公司的关注，祝您早日找到理想的工作！"})
		}
	}

	utils.Success(c, reasons)
}
