package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
		CandidateName  string
		CandidateEmail string
		JobTitle       string
		InterviewerIDs string
	}
	database.DB.QueryRow(`
		SELECT i.candidate_job_id, i.interview_time, i.method, i.location, i.link, i.interviewer_ids,
			c.name as candidate_name, c.email as candidate_email, j.title as job_title
		FROM interviews i
		LEFT JOIN candidate_jobs cj ON i.candidate_job_id = cj.id
		LEFT JOIN candidates c ON cj.candidate_id = c.id
		LEFT JOIN jobs j ON cj.job_id = j.id
		WHERE i.id = ?`,
		id,
	).Scan(&interview.CandidateJobID, &interview.InterviewTime, &interview.Method, &interview.Location, &interview.Link, &interview.InterviewerIDs,
		&interview.CandidateName, &interview.CandidateEmail, &interview.JobTitle)

	interviewTimeStr := interview.InterviewTime.Format("2006-01-02 15:04")

	var errors []string

	candidateSubject := fmt.Sprintf("面试提醒 - %s", interview.JobTitle)
	candidateBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>面试提醒</h2>
			<p>尊敬的 <strong>%s</strong>：</p>
			<p>您预约的面试信息如下：</p>
			<table style="border-collapse: collapse; margin: 15px 0;">
				<tr><td style="padding: 8px 15px; font-weight: bold;">职位：</td><td style="padding: 8px 15px;">%s</td></tr>
				<tr><td style="padding: 8px 15px; font-weight: bold;">时间：</td><td style="padding: 8px 15px;">%s</td></tr>
				<tr><td style="padding: 8px 15px; font-weight: bold;">方式：</td><td style="padding: 8px 15px;">%s</td></tr>
				%s
				%s
			</table>
			<p>请提前做好准备，准时参加面试。</p>
			<p style="color: #888; font-size: 12px;">此邮件由 HireFlow 招聘管理系统自动发送</p>
		</div>`,
		interview.CandidateName,
		interview.JobTitle,
		interviewTimeStr,
		interview.Method,
		func() string {
			if interview.Location != "" {
				return fmt.Sprintf(`<tr><td style="padding: 8px 15px; font-weight: bold;">地点：</td><td style="padding: 8px 15px;">%s</td></tr>`, interview.Location)
			}
			return ""
		}(),
		func() string {
			if interview.Link != "" {
				return fmt.Sprintf(`<tr><td style="padding: 8px 15px; font-weight: bold;">链接：</td><td style="padding: 8px 15px;"><a href="%s">%s</a></td></tr>`, interview.Link, interview.Link)
			}
			return ""
		}(),
	)

	if interview.CandidateEmail != "" {
		if err := utils.SendEmail(interview.CandidateEmail, candidateSubject, candidateBody); err != nil {
			errors = append(errors, fmt.Sprintf("候选人邮件发送失败: %v", err))
		}
	}

	if interview.InterviewerIDs != "" {
		var interviewerIDs []int
		for _, idStr := range strings.Split(interview.InterviewerIDs, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				interviewerIDs = append(interviewerIDs, id)
			}
		}
		if len(interviewerIDs) > 0 {
			placeholders := make([]string, len(interviewerIDs))
			args := make([]interface{}, len(interviewerIDs))
			for i, id := range interviewerIDs {
				placeholders[i] = "?"
				args[i] = id
			}
			rows, err := database.DB.Query("SELECT email, real_name FROM users WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
			if err == nil {
				for rows.Next() {
					var email, name string
					rows.Scan(&email, &name)
					if email == "" {
						continue
					}
					interviewerSubject := fmt.Sprintf("面试提醒 - %s - %s", interview.CandidateName, interview.JobTitle)
					interviewerBody := fmt.Sprintf(`
						<div style="font-family: Arial, sans-serif; padding: 20px;">
							<h2>面试提醒</h2>
							<p>尊敬的 <strong>%s</strong>：</p>
							<p>您有一场即将进行的面试：</p>
							<table style="border-collapse: collapse; margin: 15px 0;">
								<tr><td style="padding: 8px 15px; font-weight: bold;">候选人：</td><td style="padding: 8px 15px;">%s</td></tr>
								<tr><td style="padding: 8px 15px; font-weight: bold;">职位：</td><td style="padding: 8px 15px;">%s</td></tr>
								<tr><td style="padding: 8px 15px; font-weight: bold;">时间：</td><td style="padding: 8px 15px;">%s</td></tr>
								<tr><td style="padding: 8px 15px; font-weight: bold;">方式：</td><td style="padding: 8px 15px;">%s</td></tr>
								%s
								%s
							</table>
							<p style="color: #888; font-size: 12px;">此邮件由 HireFlow 招聘管理系统自动发送</p>
						</div>`,
						name,
						interview.CandidateName,
						interview.JobTitle,
						interviewTimeStr,
						interview.Method,
						func() string {
							if interview.Location != "" {
								return fmt.Sprintf(`<tr><td style="padding: 8px 15px; font-weight: bold;">地点：</td><td style="padding: 8px 15px;">%s</td></tr>`, interview.Location)
							}
							return ""
						}(),
						func() string {
							if interview.Link != "" {
								return fmt.Sprintf(`<tr><td style="padding: 8px 15px; font-weight: bold;">链接：</td><td style="padding: 8px 15px;"><a href="%s">%s</a></td></tr>`, interview.Link, interview.Link)
							}
							return ""
						}(),
					)
					if err := utils.SendEmail(email, interviewerSubject, interviewerBody); err != nil {
						errors = append(errors, fmt.Sprintf("%s邮件发送失败: %v", name, err))
					}
				}
				rows.Close()
			}
		}
	}

	if len(errors) > 0 {
		utils.Success(c, gin.H{"message": "部分提醒发送成功", "errors": errors})
	} else {
		utils.Success(c, gin.H{"message": "提醒已发送"})
	}
}

func GetInterviewMethods(c *gin.Context) {
	methods := []string{"现场", "视频", "电话"}
	utils.Success(c, methods)
}

func GetRecommendations(c *gin.Context) {
	recs := []string{"强烈推荐", "推荐", "待定", "不推荐"}
	utils.Success(c, recs)
}
