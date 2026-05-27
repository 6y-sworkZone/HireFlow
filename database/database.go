package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		real_name TEXT NOT NULL,
		department TEXT DEFAULT '',
		role TEXT DEFAULT 'HR',
		status TEXT DEFAULT 'active',
		avatar TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		department TEXT NOT NULL,
		location TEXT DEFAULT '',
		salary_min INTEGER DEFAULT 0,
		salary_max INTEGER DEFAULT 0,
		description TEXT DEFAULT '',
		requirements TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		status TEXT DEFAULT 'draft',
		workflow_id INTEGER DEFAULT 0,
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS job_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		title TEXT NOT NULL,
		department TEXT NOT NULL,
		location TEXT DEFAULT '',
		salary_min INTEGER DEFAULT 0,
		salary_max INTEGER DEFAULT 0,
		description TEXT DEFAULT '',
		requirements TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS candidates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT DEFAULT '',
		phone TEXT DEFAULT '',
		current_company TEXT DEFAULT '',
		work_years INTEGER DEFAULT 0,
		education TEXT DEFAULT '',
		expected_salary INTEGER DEFAULT 0,
		source TEXT DEFAULT '主动投递',
		avatar TEXT DEFAULT '',
		resume_path TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		remark TEXT DEFAULT '',
		status TEXT DEFAULT 'pool',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS candidate_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_id INTEGER NOT NULL,
		job_id INTEGER NOT NULL,
		current_stage INTEGER DEFAULT 0,
		status TEXT DEFAULT 'active',
		assigned_to INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(candidate_id, job_id),
		FOREIGN KEY (candidate_id) REFERENCES candidates(id),
		FOREIGN KEY (job_id) REFERENCES jobs(id)
	);

	CREATE TABLE IF NOT EXISTS workflows (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		stages TEXT NOT NULL,
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS stage_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_job_id INTEGER NOT NULL,
		from_stage INTEGER,
		to_stage INTEGER NOT NULL,
		operator_id INTEGER NOT NULL,
		evaluation TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (candidate_job_id) REFERENCES candidate_jobs(id),
		FOREIGN KEY (operator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS rejection_reasons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reason TEXT NOT NULL,
		template TEXT DEFAULT '',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS interviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_job_id INTEGER NOT NULL,
		interviewer_ids TEXT NOT NULL,
		interview_time DATETIME NOT NULL,
		method TEXT DEFAULT '现场',
		duration INTEGER DEFAULT 60,
		location TEXT DEFAULT '',
		link TEXT DEFAULT '',
		status TEXT DEFAULT 'scheduled',
		evaluation TEXT DEFAULT '',
		tech_score INTEGER DEFAULT 0,
		comm_score INTEGER DEFAULT 0,
		culture_score INTEGER DEFAULT 0,
		overall_score INTEGER DEFAULT 0,
		recommendation TEXT DEFAULT '',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (candidate_job_id) REFERENCES candidate_jobs(id),
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS offers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_job_id INTEGER NOT NULL,
		salary INTEGER NOT NULL,
		start_date DATE NOT NULL,
		terms TEXT DEFAULT '',
		status TEXT DEFAULT 'pending_approval',
		template_id INTEGER DEFAULT 0,
		pdf_path TEXT DEFAULT '',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (candidate_job_id) REFERENCES candidate_jobs(id),
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS offer_approval (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		offer_id INTEGER NOT NULL,
		approver_id INTEGER NOT NULL,
		"order" INTEGER NOT NULL,
		status TEXT DEFAULT 'pending',
		comment TEXT DEFAULT '',
		approved_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (offer_id) REFERENCES offers(id),
		FOREIGN KEY (approver_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS offer_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		salary_structure TEXT DEFAULT '',
		terms TEXT DEFAULT '',
		html_template TEXT DEFAULT '',
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS candidate_notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		mentions TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (candidate_id) REFERENCES candidates(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS candidate_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		candidate_id INTEGER NOT NULL,
		job_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		score INTEGER NOT NULL,
		comment TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (candidate_id) REFERENCES candidates(id),
		FOREIGN KEY (job_id) REFERENCES jobs(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT DEFAULT '',
		related_id INTEGER DEFAULT 0,
		is_read INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS weekly_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		report_date DATE NOT NULL,
		content TEXT NOT NULL,
		creator_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT DEFAULT '',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_candidates_email ON candidates(email);
	CREATE INDEX IF NOT EXISTS idx_candidates_phone ON candidates(phone);
	CREATE INDEX idx_candidate_jobs_job ON candidate_jobs(job_id);
	CREATE INDEX idx_stage_history_cj ON stage_history(candidate_job_id);
	CREATE INDEX idx_interviews_time ON interviews(interview_time);
	CREATE INDEX idx_notifications_user ON notifications(user_id);
	`

	_, err := DB.Exec(schema)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
