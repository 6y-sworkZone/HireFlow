package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"hireflow/database"
	"hireflow/middleware"
	"hireflow/models"
	"hireflow/utils"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	RealName   string `json:"real_name" binding:"required"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, password, real_name, department, role, status FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.RealName, &user.Department, &user.Role, &user.Status)

	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if user.Status != "active" {
		utils.Error(c, http.StatusForbidden, "账号已被禁用")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"real_name":  user.RealName,
			"department": user.Department,
			"role":       user.Role,
		},
	})
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", req.Username, req.Email).Scan(&exists)
	if exists > 0 {
		utils.Error(c, http.StatusConflict, "用户名或邮箱已存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "加密密码失败")
		return
	}

	role := req.Role
	if role == "" {
		role = "HR"
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, email, password, real_name, department, role) VALUES (?, ?, ?, ?, ?, ?)",
		req.Username, req.Email, string(hashedPassword), req.RealName, req.Department, role,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "注册失败")
		return
	}

	id, _ := result.LastInsertId()
	utils.Success(c, gin.H{"id": id})
}

func GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, real_name, department, role, status, avatar, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.RealName, &user.Department, &user.Role, &user.Status, &user.Avatar, &user.CreatedAt)

	if err != nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	utils.Success(c, user)
}

func GetUserList(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, username, email, real_name, department, role, status, avatar, created_at FROM users ORDER BY id",
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询用户列表失败")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.RealName, &u.Department, &u.Role, &u.Status, &u.Avatar, &u.CreatedAt)
		users = append(users, u)
	}

	utils.Success(c, users)
}

func UpdateUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		RealName   string `json:"real_name"`
		Department string `json:"department"`
		Avatar     string `json:"avatar"`
	}
	c.ShouldBindJSON(&req)

	_, err := database.DB.Exec(
		"UPDATE users SET real_name = COALESCE(NULLIF(?, ''), real_name), department = COALESCE(NULLIF(?, ''), department), avatar = COALESCE(NULLIF(?, ''), avatar), updated_at = ? WHERE id = ?",
		req.RealName, req.Department, req.Avatar, time.Now(), userID,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	utils.Success(c, nil)
}

func ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var password string
	database.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&password)
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.OldPassword)); err != nil {
		utils.Error(c, http.StatusBadRequest, "原密码错误")
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	database.DB.Exec("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", string(hashedPassword), time.Now(), userID)

	utils.Success(c, nil)
}

func EnsureDefaultUser() {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	database.DB.Exec(
		"INSERT INTO users (username, email, password, real_name, department, role) VALUES (?, ?, ?, ?, ?, ?)",
		"admin", "admin@hireflow.com", string(hashedPassword), "系统管理员", "HR", "admin",
	)
}
