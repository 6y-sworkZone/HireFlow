package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"hireflow/config"
	"hireflow/database"
	"hireflow/handlers"
	"hireflow/middleware"
)

func main() {
	cfg := config.Load()

	middleware.SetJWTSecret(cfg.JWTSecret)

	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	handlers.EnsureDefaultUser()

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.Static("/uploads", cfg.UploadDir)
	r.Static("/static", "./web/static")

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/dashboard", "./web/index.html")
	r.StaticFile("/jobs", "./web/index.html")
	r.StaticFile("/candidates", "./web/index.html")
	r.StaticFile("/workflow", "./web/index.html")
	r.StaticFile("/interviews", "./web/index.html")
	r.StaticFile("/offers", "./web/index.html")
	r.StaticFile("/analytics", "./web/index.html")
	r.StaticFile("/settings", "./web/index.html")

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
			auth.GET("/users", middleware.AuthMiddleware(), handlers.GetUserList)
			auth.PUT("/user", middleware.AuthMiddleware(), handlers.UpdateUser)
			auth.PUT("/password", middleware.AuthMiddleware(), handlers.ChangePassword)
		}

		jobs := api.Group("/jobs")
		jobs.Use(middleware.AuthMiddleware())
		{
			jobs.POST("", handlers.CreateJob)
			jobs.GET("", handlers.GetJobList)
			jobs.GET("/departments", handlers.GetDepartments)
			jobs.GET("/templates", handlers.GetJobTemplates)
			jobs.POST("/templates", handlers.CreateJobTemplate)
			jobs.DELETE("/templates/:id", handlers.DeleteJobTemplate)
			jobs.GET("/:id", handlers.GetJob)
			jobs.PUT("/:id", handlers.UpdateJob)
			jobs.DELETE("/:id", handlers.DeleteJob)
			jobs.GET("/:id/stats", handlers.GetJobStats)
		}

		candidates := api.Group("/candidates")
		candidates.Use(middleware.AuthMiddleware())
		{
			candidates.POST("", handlers.CreateCandidate)
			candidates.GET("", handlers.GetCandidateList)
			candidates.GET("/pool", handlers.GetCandidatePool)
			candidates.GET("/sources", handlers.GetSources)
			candidates.GET("/check-duplicate", handlers.CheckDuplicate)
			candidates.GET("/:id", handlers.GetCandidate)
			candidates.PUT("/:id", handlers.UpdateCandidate)
			candidates.DELETE("/:id", handlers.DeleteCandidate)
			candidates.POST("/:id/resume", handlers.UploadResume)
			candidates.GET("/:id/resume", handlers.DownloadResume)
			candidates.POST("/:id/notes", handlers.AddCandidateNote)
			candidates.GET("/:id/notes", handlers.GetCandidateNotes)
			candidates.DELETE("/:id/notes/:note_id", handlers.DeleteCandidateNote)
			candidates.POST("/:id/scores", handlers.AddCandidateScore)
			candidates.GET("/:id/scores", handlers.GetCandidateScores)
			candidates.GET("/:id/scores/summary", handlers.GetCandidateScoreSummary)
		}

		workflows := api.Group("/workflows")
		workflows.Use(middleware.AuthMiddleware())
		{
			workflows.POST("", handlers.CreateWorkflow)
			workflows.GET("", handlers.GetWorkflowList)
			workflows.GET("/:id", handlers.GetWorkflow)
			workflows.PUT("/:id", handlers.UpdateWorkflow)
			workflows.DELETE("/:id", handlers.DeleteWorkflow)
		}

		candidateJobs := api.Group("/candidate-jobs")
		candidateJobs.Use(middleware.AuthMiddleware())
		{
			candidateJobs.POST("", handlers.AssignCandidateToJob)
			candidateJobs.GET("", handlers.GetCandidateJobs)
			candidateJobs.GET("/kanban", handlers.GetKanbanView)
			candidateJobs.GET("/rejection-reasons", handlers.GetRejectionReasons)
			candidateJobs.GET("/:id", handlers.GetCandidateJob)
			candidateJobs.PUT("/:id/move-stage", handlers.MoveStage)
			candidateJobs.PUT("/:id/reject", handlers.RejectCandidate)
			candidateJobs.GET("/:id/history", handlers.GetStageHistory)
		}

		interviews := api.Group("/interviews")
		interviews.Use(middleware.AuthMiddleware())
		{
			interviews.POST("", handlers.CreateInterview)
			interviews.GET("", handlers.GetInterviewList)
			interviews.GET("/calendar", handlers.GetInterviewCalendar)
			interviews.GET("/methods", handlers.GetInterviewMethods)
			interviews.GET("/recommendations", handlers.GetRecommendations)
			interviews.GET("/:id", handlers.GetInterview)
			interviews.PUT("/:id", handlers.UpdateInterview)
			interviews.PUT("/:id/evaluate", handlers.SubmitInterviewEvaluation)
			interviews.POST("/:id/remind", handlers.SendInterviewReminder)
			interviews.DELETE("/:id", handlers.DeleteInterview)
		}

		offers := api.Group("/offers")
		offers.Use(middleware.AuthMiddleware())
		{
			offers.POST("", handlers.CreateOffer)
			offers.GET("", handlers.GetOfferList)
			offers.GET("/templates", handlers.GetOfferTemplates)
			offers.POST("/templates", handlers.CreateOfferTemplate)
			offers.DELETE("/templates/:id", handlers.DeleteOfferTemplate)
			offers.GET("/statuses", handlers.GetOfferStatuses)
			offers.GET("/:id", handlers.GetOffer)
			offers.PUT("/:id", handlers.UpdateOffer)
			offers.DELETE("/:id", handlers.DeleteOffer)
			offers.POST("/:id/approval", handlers.CreateOfferApproval)
			offers.GET("/:id/approval", handlers.GetOfferApproval)
			offers.PUT("/:id/approve", handlers.ApproveOffer)
			offers.PUT("/:id/reject", handlers.RejectOffer)
			offers.POST("/:id/send", handlers.SendOffer)
			offers.PUT("/:id/status", handlers.UpdateOfferStatus)
			offers.GET("/:id/pdf", handlers.DownloadOfferPDF)
		}

		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware())
		{
			analytics.GET("/funnel", handlers.GetRecruitmentFunnel)
			analytics.GET("/jobs", handlers.GetJobStatsByDimension)
			analytics.GET("/channels", handlers.GetChannelStats)
			analytics.GET("/interviewers", handlers.GetInterviewerStats)
			analytics.GET("/monthly-trend", handlers.GetMonthlyTrend)
			analytics.GET("/department-progress", handlers.GetDepartmentProgress)
			analytics.GET("/offer-acceptance", handlers.GetOfferAcceptanceRate)
			analytics.GET("/stage-stay", handlers.GetStageStayStats)
		}

		settings := api.Group("/settings")
		settings.Use(middleware.AuthMiddleware())
		{
			settings.GET("", handlers.GetSettings)
			settings.POST("", handlers.SaveSettings)
			settings.POST("/test-smtp", handlers.TestSMTPConnection)
		}

		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			notifications.GET("", handlers.GetNotifications)
			notifications.GET("/unread-count", handlers.GetUnreadCount)
			notifications.PUT("/:id/read", handlers.ReadNotification)
			notifications.PUT("/read-all", handlers.ReadAllNotifications)
		}

		reports := api.Group("/reports")
		reports.Use(middleware.AuthMiddleware())
		{
			reports.POST("/weekly", handlers.GenerateWeeklyReport)
			reports.GET("/weekly", handlers.GetWeeklyReports)
		}

		mentions := api.Group("/mentions")
		mentions.Use(middleware.AuthMiddleware())
		{
			mentions.GET("/users", handlers.GetUserMentionList)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("HireFlow server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
