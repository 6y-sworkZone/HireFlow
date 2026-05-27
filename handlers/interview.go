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

func CreateInterview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var interview models.Interview
	c.ShouldBindJSON(&interview)

	if interview.Status == "" {
		interview.Status = "scheduled"
	}
	if interview.Duration == 0 {
		interview.Duration = 60
	}

	result, err := database.DB.Exec(
		`INSERT INTO interviews (candidate_job_id, interviewer_ids, interview_time, method, duration, 
		 location, link, status, creator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		interview.CandidateJobID, interview.InterviewerIDs, interview.InterviewTime,
		interview.Method, interview.Duration, interview.Location, interview.Link,
		interview.Status, userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建面试失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetInterviewList(c *gin.Context) {
	interviewerID := c.Query("interviewer_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	where := "WHERE 1=1"
	args := []interface{}{}

	if interviewerID != "" {
		where += " AND i.interviewer_ids LIKE ?"
		args = append(args, "%"+interviewerID+"%")
	}
	if startDate != "" {
		where += " AND DATE(i.interview_time) >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		where += " AND DATE(i.interview_time) <= ?"
		args = append(args, endDate)
	}
	if status != "" {
		where += " AND i.status = ?"
		args = append(args, status)
	}

	query := `SELECT i.id, i.candidate_job_id, i.interviewer_ids, i.interview_time, i.method,
			  i.duration, i.location, i.link, i.status, i.evaluation, i.tech_score, i.comm_score,
			  i.culture_score, i.overall_score, i.recommendation, i.created_at, i.updated_at,
			  c.name as candidate_name, j.title as job_title
			  FROM interviews i
			  LEFT JOIN candidate_jobs cj ON i.candidate_job_id = cj.id
			  LEFT JOIN candidates c ON cj.candidate_id = c.id
			  LEFT JOIN jobs j ON cj.job_id = j.id ` + where + ` ORDER BY i.interview_time DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询面试列表失败")
		return
	}
	defer rows.Close()

	var interviews []models.Interview
	for rows.Next() {
		var itv models.Interview
		rows.Scan(&itv.ID, &itv.CandidateJobID, &itv.InterviewerIDs, &itv.InterviewTime,
			&itv.Method, &itv.Duration, &itv.Location, &itv.Link, &itv.Status, &itv.Evaluation,
			&itv.TechScore, &itv.CommScore, &itv.CultureScore, &itv.OverallScore,
			&itv.Recommendation, &itv.CreatedAt, &itv.UpdatedAt, &itv.CandidateName, &itv.JobTitle)
		interviews = append(interviews, itv)
	}

	utils.Success(c, interviews)
}

func GetInterview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var itv models.Interview
	err := database.DB.QueryRow(
		`SELECT i.id, i.candidate_job_id, i.interviewer_ids, i.interview_time, i.method,
		 i.duration, i.location, i.link, i.status, i.evaluation, i.tech_score, i.comm_score,
		 i.culture_score, i.overall_score, i.recommendation, i.creator_id, i.created_at, i.updated_at,
		 c.name as candidate_name, j.title as job_title, cj.current_stage
		 FROM interviews i
		 LEFT JOIN candidate_jobs cj ON i.candidate_job_id = cj.id
		 LEFT JOIN candidates c ON cj.candidate_id = c.id
		 LEFT JOIN jobs j ON cj.job_id = j.id WHERE i.id = ?`,
		id,
	).Scan(&itv.ID, &itv.CandidateJobID, &itv.InterviewerIDs, &itv.InterviewTime,
		&itv.Method, &itv.Duration, &itv.Location, &itv.Link, &itv.Status, &itv.Evaluation,
		&itv.TechScore, &itv.CommScore, &itv.CultureScore, &itv.OverallScore,
		&itv.Recommendation, &itv.CreatorID, &itv.CreatedAt, &itv.UpdatedAt,
		&itv.CandidateName, &itv.JobTitle, &itv.CandidateJobID)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "面试不存在")
		return
	}

	utils.Success(c, itv)
}

func UpdateInterview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		InterviewerIDs string    `json:"interviewer_ids"`
		InterviewTime  time.Time `json:"interview_time"`
		Method         string    `json:"method"`
		Duration       int       `json:"duration"`
		Location       string    `json:"location"`
		Link           string    `json:"link"`
		Status         string    `json:"status"`
	}
	c.ShouldBindJSON(&req)

	_, err := database.DB.Exec(
		`UPDATE interviews SET interviewer_ids = COALESCE(NULLIF(?, ''), interviewer_ids),
		 interview_time = COALESCE(?, interview_time), method = COALESCE(NULLIF(?, ''), method),
		 duration = COALESCE(NULLIF(?, 0), duration), location = ?, link = ?,
		 status = COALESCE(NULLIF(?, ''), status), updated_at = ? WHERE id = ?`,
		req.InterviewerIDs, req.InterviewTime, req.Method, req.Duration,
		req.Location, req.Link, req.Status, time.Now(), id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新面试失败")
		return
	}

	utils.Success(c, nil)
}

func SubmitInterviewEvaluation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Evaluation     string `json:"evaluation"`
		TechScore      int    `json:"tech_score"`
		CommScore      int    `json:"comm_score"`
		CultureScore   int    `json:"culture_score"`
		OverallScore   int    `json:"overall_score"`
		Recommendation string `json:"recommendation"`
	}
	c.ShouldBindJSON(&req)

	_, err := database.DB.Exec(
		`UPDATE interviews SET evaluation = ?, tech_score = ?, comm_score = ?, culture_score = ?,
		 overall_score = ?, recommendation = ?, status = 'completed', updated_at = ? WHERE id = ?`,
		req.Evaluation, req.TechScore, req.CommScore, req.CultureScore,
		req.OverallScore, req.Recommendation, time.Now(), id,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "提交评价失败")
		return
	}

	_ = userID

	utils.Success(c, nil)
}

func DeleteInterview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM interviews WHERE id = ?", id)
	utils.Success(c, nil)
}

func GetInterviewCalendar(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	interviewerID := c.Query("interviewer_id")

	where := "WHERE DATE(i.interview_time) >= ? AND DATE(i.interview_time) <= ?"
	args := []interface{}{startDate, endDate}

	if interviewerID != "" {
		where += " AND i.interviewer_ids LIKE ?"
		args = append(args, "%"+interviewerID+"%")
	}

	query := `SELECT i.id, i.candidate_job_id, i.interviewer_ids, i.interview_time, i.method,
			  i.duration, i.location, i.link, i.status,
			  c.name as candidate_name, j.title as job_title
			  FROM interviews i
			  LEFT JOIN candidate_jobs cj ON i.candidate_job_id = cj.id
			  LEFT JOIN candidates c ON cj.candidate_id = c.id
			  LEFT JOIN jobs j ON cj.job_id = j.id ` + where + ` ORDER BY i.interview_time`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询日历失败")
		return
	}
	defer rows.Close()

	type CalendarEvent struct {
		ID              int       `json:"id"`
		CandidateJobID  int       `json:"candidate_job_id"`
		InterviewerIDs  string    `json:"interviewer_ids"`
		InterviewTime   time.Time `json:"interview_time"`
		Method          string    `json:"method"`
		Duration        int       `json:"duration"`
		Location        string    `json:"location"`
		Link            string    `json:"link"`
		Status          string    `json:"status"`
		CandidateName   string    `json:"candidate_name"`
		JobTitle        string    `json:"job_title"`
	}

	var events []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		rows.Scan(&e.ID, &e.CandidateJobID, &e.InterviewerIDs, &e.InterviewTime, &e.Method,
			&e.Duration, &e.Location, &e.Link, &e.Status, &e.CandidateName, &e.JobTitle)
		events = append(events, e)
	}

	utils.Success(c, events)
}

func SendInterviewReminder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var interview struct {
		CandidateJobID int
		InterviewTime  time.Time
		Method         string
		Location       string
		Link           string
	}
	database.DB.QueryRow(
		"SELECT candidate_job_id, interview_time, method, location, link FROM interviews WHERE id = ?",
		id,
	).Scan(&interview.CandidateJobID, &interview.InterviewTime, &interview.Method, &interview.Location, &interview.Link)

	utils.Success(c, gin.H{"message": "提醒已发送"})
}

func GetInterviewMethods(c *gin.Context) {
	methods := []string{"现场", "视频", "电话"}
	utils.Success(c, methods)
}

func GetRecommendations(c *gin.Context) {
	recs := []string{"强烈推荐", "推荐", "待定", "不推荐"}
	utils.Success(c, recs)
}
