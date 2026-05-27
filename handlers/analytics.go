package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hireflow/database"
	"hireflow/utils"
)

func GetRecruitmentFunnel(c *gin.Context) {
	jobID := c.Query("job_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	where := "WHERE 1=1"
	args := []interface{}{}

	if jobID != "" {
		where += " AND cj.job_id = ?"
		args = append(args, jobID)
	}
	if startDate != "" {
		where += " AND DATE(cj.created_at) >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		where += " AND DATE(cj.created_at) <= ?"
		args = append(args, endDate)
	}

	type FunnelStage struct {
		Stage      int    `json:"stage"`
		StageName  string `json:"stage_name"`
		Count      int    `json:"count"`
		Rate       float64 `json:"rate"`
	}

	stages := []FunnelStage{
		{Stage: 1, StageName: "简历筛选"},
		{Stage: 2, StageName: "电话面试"},
		{Stage: 3, StageName: "技术面试"},
		{Stage: 4, StageName: "HR面试"},
		{Stage: 5, StageName: "Offer"},
		{Stage: 6, StageName: "入职"},
	}

	var totalResumes int
	database.DB.QueryRow("SELECT COUNT(*) FROM candidate_jobs cj "+where, args...).Scan(&totalResumes)

	if totalResumes == 0 {
		totalResumes = 1
	}

	for i := range stages {
		stageWhere := where + " AND cj.current_stage >= ?"
		stageArgs := append([]interface{}{}, args...)
		stageArgs = append(stageArgs, stages[i].Stage)

		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM candidate_jobs cj "+stageWhere, stageArgs...).Scan(&count)
		stages[i].Count = count
		stages[i].Rate = float64(count) / float64(totalResumes) * 100
	}

	utils.Success(c, gin.H{
		"stages":       stages,
		"total_resumes": totalResumes,
	})
}

func GetJobStatsByDimension(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT j.id, j.title, j.department, j.status,
			COUNT(DISTINCT cj.id) as candidate_count,
			COUNT(DISTINCT CASE WHEN i.id IS NOT NULL THEN cj.id END) as interview_count,
			COUNT(DISTINCT CASE WHEN o.id IS NOT NULL THEN cj.id END) as offer_count,
			AVG(CASE WHEN cj.current_stage >= 6 THEN 
				CAST((julianday(cj.updated_at) - julianday(cj.created_at)) * 24 AS INTEGER) 
			END) as avg_days
		FROM jobs j
		LEFT JOIN candidate_jobs cj ON cj.job_id = j.id
		LEFT JOIN interviews i ON i.candidate_job_id = cj.id
		LEFT JOIN offers o ON o.candidate_job_id = cj.id
		GROUP BY j.id
		ORDER BY j.created_at DESC
	`)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询职位统计失败")
		return
	}
	defer rows.Close()

	type JobStat struct {
		JobID          int     `json:"job_id"`
		Title          string  `json:"title"`
		Department     string  `json:"department"`
		Status         string  `json:"status"`
		CandidateCount int     `json:"candidate_count"`
		InterviewCount int     `json:"interview_count"`
		OfferCount     int     `json:"offer_count"`
		AvgDays        float64 `json:"avg_days"`
	}

	var stats []JobStat
	for rows.Next() {
		var s JobStat
		rows.Scan(&s.JobID, &s.Title, &s.Department, &s.Status,
			&s.CandidateCount, &s.InterviewCount, &s.OfferCount, &s.AvgDays)
		stats = append(stats, s)
	}

	utils.Success(c, stats)
}

func GetChannelStats(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT c.source,
			COUNT(*) as total_count,
			COUNT(DISTINCT CASE WHEN cj.id IS NOT NULL THEN c.id END) as active_count,
			COUNT(DISTINCT CASE WHEN o.id IS NOT NULL THEN c.id END) as offer_count,
			COUNT(DISTINCT CASE WHEN cj.current_stage >= 6 THEN c.id END) as hired_count
		FROM candidates c
		LEFT JOIN candidate_jobs cj ON cj.candidate_id = c.id
		LEFT JOIN offers o ON o.candidate_job_id = cj.id
		GROUP BY c.source
		ORDER BY total_count DESC
	`)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询渠道统计失败")
		return
	}
	defer rows.Close()

	type ChannelStat struct {
		Source      string  `json:"source"`
		TotalCount  int     `json:"total_count"`
		ActiveCount int     `json:"active_count"`
		OfferCount  int     `json:"offer_count"`
		HiredCount  int     `json:"hired_count"`
		HireRate    float64 `json:"hire_rate"`
	}

	var stats []ChannelStat
	for rows.Next() {
		var s ChannelStat
		rows.Scan(&s.Source, &s.TotalCount, &s.ActiveCount, &s.OfferCount, &s.HiredCount)
		if s.TotalCount > 0 {
			s.HireRate = float64(s.HiredCount) / float64(s.TotalCount) * 100
		}
		stats = append(stats, s)
	}

	utils.Success(c, stats)
}

func GetInterviewerStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	where := "WHERE 1=1"
	args := []interface{}{}

	if startDate != "" {
		where += " AND DATE(i.interview_time) >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		where += " AND DATE(i.interview_time) <= ?"
		args = append(args, endDate)
	}

	query := `SELECT u.id, u.real_name, u.department,
			  COUNT(*) as interview_count,
			  AVG(i.overall_score) as avg_score
			  FROM interviews i
			  JOIN users u ON i.interviewer_ids LIKE '%' || u.id || '%'
			  ` + where + `
			  GROUP BY u.id
			  ORDER BY interview_count DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询面试官统计失败")
		return
	}
	defer rows.Close()

	type InterviewerStat struct {
		UserID          int     `json:"user_id"`
		RealName        string  `json:"real_name"`
		Department      string  `json:"department"`
		InterviewCount  int     `json:"interview_count"`
		AvgScore        float64 `json:"avg_score"`
	}

	var stats []InterviewerStat
	for rows.Next() {
		var s InterviewerStat
		rows.Scan(&s.UserID, &s.RealName, &s.Department, &s.InterviewCount, &s.AvgScore)
		stats = append(stats, s)
	}

	utils.Success(c, stats)
}

func GetMonthlyTrend(c *gin.Context) {
	months := []string{}
	now := time.Now()
	for i := 5; i >= 0; i-- {
		month := now.AddDate(0, -i, 0)
		months = append(months, month.Format("2006-01"))
	}

	type MonthlyData struct {
		Month           string `json:"month"`
		NewCandidates   int    `json:"new_candidates"`
		Interviews      int    `json:"interviews"`
		Offers          int    `json:"offers"`
		Hired           int    `json:"hired"`
	}

	var trend []MonthlyData
	for _, month := range months {
		var item MonthlyData
		item.Month = month

		database.DB.QueryRow(
			"SELECT COUNT(*) FROM candidates WHERE strftime('%Y-%m', created_at) = ?",
			month,
		).Scan(&item.NewCandidates)

		database.DB.QueryRow(
			"SELECT COUNT(*) FROM interviews WHERE strftime('%Y-%m', interview_time) = ?",
			month,
		).Scan(&item.Interviews)

		database.DB.QueryRow(
			"SELECT COUNT(*) FROM offers WHERE strftime('%Y-%m', created_at) = ? AND status IN ('sent', 'accepted')",
			month,
		).Scan(&item.Offers)

		database.DB.QueryRow(`
			SELECT COUNT(*) FROM candidate_jobs cj
			WHERE strftime('%Y-%m', updated_at) = ? AND current_stage >= 6
		`, month).Scan(&item.Hired)

		trend = append(trend, item)
	}

	utils.Success(c, trend)
}

func GetDepartmentProgress(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT j.department,
			COUNT(*) as total_jobs,
			COUNT(DISTINCT cj.id) as total_candidates,
			COUNT(DISTINCT CASE WHEN o.id IS NOT NULL THEN cj.id END) as total_offers,
			COUNT(DISTINCT CASE WHEN cj.current_stage >= 6 THEN cj.id END) as total_hired
		FROM jobs j
		LEFT JOIN candidate_jobs cj ON cj.job_id = j.id
		LEFT JOIN offers o ON o.candidate_job_id = cj.id
		GROUP BY j.department
		ORDER BY total_candidates DESC
	`)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询部门进度失败")
		return
	}
	defer rows.Close()

	type DepartmentProgress struct {
		Department       string `json:"department"`
		TotalJobs        int    `json:"total_jobs"`
		TotalCandidates  int    `json:"total_candidates"`
		TotalOffers      int    `json:"total_offers"`
		TotalHired       int    `json:"total_hired"`
	}

	var progress []DepartmentProgress
	for rows.Next() {
		var p DepartmentProgress
		rows.Scan(&p.Department, &p.TotalJobs, &p.TotalCandidates, &p.TotalOffers, &p.TotalHired)
		progress = append(progress, p)
	}

	utils.Success(c, progress)
}

func GetOfferAcceptanceRate(c *gin.Context) {
	var totalSent, totalAccepted int

	database.DB.QueryRow("SELECT COUNT(*) FROM offers WHERE status = 'sent'").Scan(&totalSent)
	database.DB.QueryRow("SELECT COUNT(*) FROM offers WHERE status = 'accepted'").Scan(&totalAccepted)

	var rate float64
	if totalSent > 0 {
		rate = float64(totalAccepted) / float64(totalSent) * 100
	}

	utils.Success(c, gin.H{
		"total_sent":     totalSent,
		"total_accepted": totalAccepted,
		"acceptance_rate": rate,
	})
}

func GetStageStayStats(c *gin.Context) {
	jobID := c.Query("job_id")

	where := "WHERE sh.from_stage != 0"
	args := []interface{}{}

	if jobID != "" {
		where += " AND cj.job_id = ?"
		args = append(args, jobID)
	}

	rows, err := database.DB.Query(`
		SELECT sh.from_stage, 
			AVG(julianday(sh.created_at) - julianday(prev_sh.created_at)) as avg_days
		FROM stage_history sh
		JOIN candidate_jobs cj ON sh.candidate_job_id = cj.id
		LEFT JOIN stage_history prev_sh ON prev_sh.candidate_job_id = sh.candidate_job_id 
			AND prev_sh.to_stage = sh.from_stage
		`+where+`
		GROUP BY sh.from_stage
		ORDER BY sh.from_stage
	`, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询阶段停留时长失败")
		return
	}
	defer rows.Close()

	type StageStay struct {
		Stage   int     `json:"stage"`
		AvgDays float64 `json:"avg_days"`
	}

	stageNames := map[int]string{
		1: "简历筛选",
		2: "电话面试",
		3: "技术面试",
		4: "HR面试",
		5: "Offer",
		6: "入职",
	}

	var stats []map[string]interface{}
	for rows.Next() {
		var s StageStay
		rows.Scan(&s.Stage, &s.AvgDays)
		stats = append(stats, map[string]interface{}{
			"stage":      s.Stage,
			"stage_name": stageNames[s.Stage],
			"avg_days":   s.AvgDays,
		})
	}

	utils.Success(c, stats)
}
