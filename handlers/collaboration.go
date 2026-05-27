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

func AddCandidateNote(c *gin.Context) {
	userID := middleware.GetUserID(c)
	candidateID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Content  string `json:"content" binding:"required"`
		Mentions string `json:"mentions"`
	}
	c.ShouldBindJSON(&req)

	result, err := database.DB.Exec(
		"INSERT INTO candidate_notes (candidate_id, user_id, content, mentions) VALUES (?, ?, ?, ?)",
		candidateID, userID, req.Content, req.Mentions,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "添加备注失败")
		return
	}

	noteID, _ := result.LastInsertId()

	if req.Mentions != "" {
		mentionIDs := strings.Split(req.Mentions, ",")
		for _, mid := range mentionIDs {
			uid, _ := strconv.Atoi(mid)
			if uid > 0 {
				database.DB.Exec(
					"INSERT INTO notifications (user_id, type, title, content, related_id) VALUES (?, 'mention', ?, ?, ?)",
					uid, fmt.Sprintf("有人在候选人备注中@了你"), req.Content, candidateID,
				)
			}
		}
	}

	utils.Success(c, gin.H{"id": noteID})
}

func GetCandidateNotes(c *gin.Context) {
	candidateID, _ := strconv.Atoi(c.Param("id"))

	rows, err := database.DB.Query(
		`SELECT cn.id, cn.candidate_id, cn.user_id, cn.content, cn.mentions, cn.created_at,
		 u.real_name as user_name
		 FROM candidate_notes cn
		 LEFT JOIN users u ON cn.user_id = u.id
		 WHERE cn.candidate_id = ?
		 ORDER BY cn.created_at DESC`,
		candidateID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询备注列表失败")
		return
	}
	defer rows.Close()

	var notes []models.CandidateNote
	for rows.Next() {
		var n models.CandidateNote
		rows.Scan(&n.ID, &n.CandidateID, &n.UserID, &n.Content, &n.Mentions, &n.CreatedAt, &n.UserName)
		notes = append(notes, n)
	}

	utils.Success(c, notes)
}

func DeleteCandidateNote(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("note_id"))
	database.DB.Exec("DELETE FROM candidate_notes WHERE id = ?", id)
	utils.Success(c, nil)
}

func AddCandidateScore(c *gin.Context) {
	userID := middleware.GetUserID(c)
	candidateID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		JobID   int    `json:"job_id" binding:"required"`
		Score   int    `json:"score" binding:"required"`
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&req)

	result, err := database.DB.Exec(
		"INSERT INTO candidate_scores (candidate_id, job_id, user_id, score, comment) VALUES (?, ?, ?, ?, ?)",
		candidateID, req.JobID, userID, req.Score, req.Comment,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "评分失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetCandidateScores(c *gin.Context) {
	candidateID, _ := strconv.Atoi(c.Param("id"))
	jobID := c.Query("job_id")

	where := "WHERE cs.candidate_id = ?"
	args := []interface{}{candidateID}

	if jobID != "" {
		where += " AND cs.job_id = ?"
		args = append(args, jobID)
	}

	rows, err := database.DB.Query(
		`SELECT cs.id, cs.candidate_id, cs.job_id, cs.user_id, cs.score, cs.comment, cs.created_at,
		 u.real_name as user_name
		 FROM candidate_scores cs
		 LEFT JOIN users u ON cs.user_id = u.id
		 `+where+`
		 ORDER BY cs.created_at DESC`,
		args...,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询评分列表失败")
		return
	}
	defer rows.Close()

	var scores []models.CandidateScore
	for rows.Next() {
		var s models.CandidateScore
		rows.Scan(&s.ID, &s.CandidateID, &s.JobID, &s.UserID, &s.Score, &s.Comment, &s.CreatedAt, &s.UserName)
		scores = append(scores, s)
	}

	utils.Success(c, scores)
}

func GetCandidateScoreSummary(c *gin.Context) {
	candidateID, _ := strconv.Atoi(c.Param("id"))
	jobID := c.Query("job_id")

	var totalScores, avgScore, count int

	query := "SELECT COUNT(*), COALESCE(SUM(score), 0) FROM candidate_scores WHERE candidate_id = ?"
	args := []interface{}{candidateID}

	if jobID != "" {
		query += " AND job_id = ?"
		args = append(args, jobID)
	}

	database.DB.QueryRow(query, args...).Scan(&count, &totalScores)

	if count > 0 {
		avgScore = totalScores / count
	}

	utils.Success(c, gin.H{
		"count":     count,
		"avg_score": avgScore,
		"total":     totalScores,
	})
}

func GetNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)

	rows, err := database.DB.Query(
		"SELECT id, user_id, type, title, content, related_id, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT 50",
		userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询通知失败")
		return
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.RelatedID, &n.IsRead, &n.CreatedAt)
		notifications = append(notifications, n)
	}

	utils.Success(c, notifications)
}

func ReadNotification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))

	database.DB.Exec("UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?", id, userID)
	utils.Success(c, nil)
}

func ReadAllNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)

	database.DB.Exec("UPDATE notifications SET is_read = 1 WHERE user_id = ?", userID)
	utils.Success(c, nil)
}

func GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0", userID).Scan(&count)

	utils.Success(c, gin.H{"unread_count": count})
}

func GenerateWeeklyReport(c *gin.Context) {
	userID := middleware.GetUserID(c)

	now := timeNow()
	weekStart := now.AddDate(0, 0, -6)
	weekStartStr := weekStart.Format("2006-01-02")
	reportDate := now.Format("2006-01-02")

	var newCandidates int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM candidates WHERE DATE(created_at) >= ?",
		weekStartStr,
	).Scan(&newCandidates)

	var interviews int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM interviews WHERE DATE(interview_time) >= ?",
		weekStartStr,
	).Scan(&interviews)

	var offers int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM offers WHERE DATE(created_at) >= ? AND status IN ('sent', 'accepted')",
		weekStartStr,
	).Scan(&offers)

	var hired int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM candidate_jobs WHERE DATE(updated_at) >= ? AND current_stage >= 6",
		weekStartStr,
	).Scan(&hired)

	content := fmt.Sprintf(`
## 本周招聘周报（%s ~ %s）

### 数据概览
- 新增简历：%d 份
- 面试安排：%d 场
- Offer发放：%d 个
- 成功入职：%d 人

### 本周重点
1. 新增简历 %d 份，较上周有所变化
2. 完成面试 %d 场
3. 发出 Offer %d 个
4. 入职 %d 人

### 下周计划
请各面试官合理安排时间，确保面试进度。
`, weekStartStr, reportDate, newCandidates, interviews, offers, hired, newCandidates, interviews, offers, hired)

	result, _ := database.DB.Exec(
		"INSERT INTO weekly_reports (report_date, content, creator_id) VALUES (?, ?, ?)",
		reportDate, content, userID,
	)

	id, _ := result.LastInsertId()

	utils.Success(c, gin.H{
		"id":      id,
		"content": content,
		"date":    reportDate,
	})
}

func GetWeeklyReports(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, report_date, content, creator_id, created_at FROM weekly_reports ORDER BY report_date DESC LIMIT 12",
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询周报失败")
		return
	}
	defer rows.Close()

	var reports []models.WeeklyReport
	for rows.Next() {
		var r models.WeeklyReport
		rows.Scan(&r.ID, &r.ReportDate, &r.Content, &r.CreatorID, &r.CreatedAt)
		reports = append(reports, r)
	}

	utils.Success(c, reports)
}

func GetUserMentionList(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, real_name, department FROM users WHERE status = 'active' ORDER BY department, real_name",
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询用户列表失败")
		return
	}
	defer rows.Close()

	type UserMention struct {
		ID         int    `json:"id"`
		RealName   string `json:"real_name"`
		Department string `json:"department"`
	}

	var users []UserMention
	for rows.Next() {
		var u UserMention
		rows.Scan(&u.ID, &u.RealName, &u.Department)
		users = append(users, u)
	}

	utils.Success(c, users)
}

func timeNow() time.Time {
	return time.Now()
}
