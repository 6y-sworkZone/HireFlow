package models

import "time"

type User struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	RealName   string    `json:"real_name"`
	Department string    `json:"department"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	Avatar     string    `json:"avatar"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Job struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Department  string    `json:"department"`
	Location    string    `json:"location"`
	SalaryMin   int       `json:"salary_min"`
	SalaryMax   int       `json:"salary_max"`
	Description string    `json:"description"`
	Requirements string   `json:"requirements"`
	Tags        string    `json:"tags"`
	Status      string    `json:"status"`
	WorkflowID  int       `json:"workflow_id"`
	CreatorID   int       `json:"creator_id"`
	CreatorName string    `json:"creator_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobTemplate struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Department   string    `json:"department"`
	Location     string    `json:"location"`
	SalaryMin    int       `json:"salary_min"`
	SalaryMax    int       `json:"salary_max"`
	Description  string    `json:"description"`
	Requirements string    `json:"requirements"`
	Tags         string    `json:"tags"`
	CreatorID    int       `json:"creator_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type Candidate struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	CurrentCompany  string    `json:"current_company"`
	WorkYears       int       `json:"work_years"`
	Education       string    `json:"education"`
	ExpectedSalary  int       `json:"expected_salary"`
	Source          string    `json:"source"`
	Avatar          string    `json:"avatar"`
	ResumePath      string    `json:"resume_path"`
	Tags            string    `json:"tags"`
	Remark          string    `json:"remark"`
	Status          string    `json:"status"`
	CreatorID       int       `json:"creator_id"`
	CreatorName     string    `json:"creator_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CandidateJob struct {
	ID            int       `json:"id"`
	CandidateID   int       `json:"candidate_id"`
	JobID         int       `json:"job_id"`
	CurrentStage  int       `json:"current_stage"`
	Status        string    `json:"status"`
	AssignedTo    int       `json:"assigned_to"`
	CandidateName string    `json:"candidate_name,omitempty"`
	CandidateInfo *Candidate `json:"candidate,omitempty"`
	JobTitle      string    `json:"job_title,omitempty"`
	JobInfo       *Job      `json:"job,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Workflow struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Stages      string    `json:"stages"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type StageHistory struct {
	ID             int       `json:"id"`
	CandidateJobID int       `json:"candidate_job_id"`
	FromStage      int       `json:"from_stage"`
	ToStage        int       `json:"to_stage"`
	OperatorID     int       `json:"operator_id"`
	OperatorName   string    `json:"operator_name,omitempty"`
	Evaluation     string    `json:"evaluation"`
	CreatedAt      time.Time `json:"created_at"`
}

type Interview struct {
	ID              int       `json:"id"`
	CandidateJobID  int       `json:"candidate_job_id"`
	InterviewerIDs  string    `json:"interviewer_ids"`
	Interviewers    []User    `json:"interviewers,omitempty"`
	InterviewTime   time.Time `json:"interview_time"`
	Method          string    `json:"method"`
	Duration        int       `json:"duration"`
	Location        string    `json:"location"`
	Link            string    `json:"link"`
	Status          string    `json:"status"`
	Evaluation      string    `json:"evaluation"`
	TechScore       int       `json:"tech_score"`
	CommScore       int       `json:"comm_score"`
	CultureScore    int       `json:"culture_score"`
	OverallScore    int       `json:"overall_score"`
	Recommendation  string    `json:"recommendation"`
	CreatorID       int       `json:"creator_id"`
	CandidateName   string    `json:"candidate_name,omitempty"`
	JobTitle        string    `json:"job_title,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Offer struct {
	ID             int       `json:"id"`
	CandidateJobID int       `json:"candidate_job_id"`
	Salary         int       `json:"salary"`
	StartDate      string    `json:"start_date"`
	Terms          string    `json:"terms"`
	Status         string    `json:"status"`
	TemplateID     int       `json:"template_id"`
	PDFPath        string    `json:"pdf_path"`
	CreatorID      int       `json:"creator_id"`
	CandidateName  string    `json:"candidate_name,omitempty"`
	JobTitle       string    `json:"job_title,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OfferApproval struct {
	ID         int        `json:"id"`
	OfferID    int        `json:"offer_id"`
	ApproverID int        `json:"approver_id"`
	ApproverName string  `json:"approver_name,omitempty"`
	Order      int        `json:"order"`
	Status     string     `json:"status"`
	Comment    string     `json:"comment"`
	ApprovedAt *time.Time `json:"approved_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type OfferTemplate struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	SalaryStructure string    `json:"salary_structure"`
	Terms           string    `json:"terms"`
	HTMLTemplate    string    `json:"html_template"`
	CreatorID       int       `json:"creator_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type CandidateNote struct {
	ID        int       `json:"id"`
	CandidateID int     `json:"candidate_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Content   string    `json:"content"`
	Mentions  string    `json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
}

type CandidateScore struct {
	ID          int       `json:"id"`
	CandidateID int       `json:"candidate_id"`
	JobID       int       `json:"job_id"`
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	Score       int       `json:"score"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	RelatedID int       `json:"related_id"`
	IsRead    int       `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type WeeklyReport struct {
	ID        int       `json:"id"`
	ReportDate string   `json:"report_date"`
	Content   string    `json:"content"`
	CreatorID int       `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RejectionReason struct {
	ID        int       `json:"id"`
	Reason    string    `json:"reason"`
	Template  string    `json:"template"`
	CreatorID int       `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

type JobStats struct {
	JobID       int `json:"job_id"`
	ResumeCount int `json:"resume_count"`
	InterviewCount int `json:"interview_count"`
	OfferCount  int `json:"offer_count"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	From     string `json:"from"`
	Security string `json:"security"`
}
