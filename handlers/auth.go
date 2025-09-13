package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/models"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CompanyID   uint   `json:"companyId"`
	CompanyName string `json:"companyName"`
}

type Claims struct {
	UserId    uint   `json:"userId"`
	Role      string `json:"role"`
	CompanyID uint   `json:"companyId"`
	jwt.RegisteredClaims
}

// POST /signup
func Signup(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existing models.User
	if err := database.DB.Where("email = ?", creds.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	var user models.User

	// CASE 1: No CompanyID provided → Create new company and assign user as admin
	if creds.CompanyID == 0 {
		companyName := creds.CompanyName
		if companyName == "" {
			companyName = creds.Username + "'s Company"
		}
		company := models.Company{Name: companyName}

		if err := database.DB.Create(&company).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create company"})
			return
		}

		user = models.User{
			Username:  creds.Username,
			Email:     creds.Email,
			Password:  string(hashedPassword),
			Role:      "admin", // first user of company is admin
			CompanyID: company.ID,
			Company:   company,
			Verified:  true, // admins are always verified
		}

	} else {
		// CASE 2: Attach to existing company as normal user
		user = models.User{
			Username:  creds.Username,
			Email:     creds.Email,
			Password:  string(hashedPassword),
			Role:      "user",
			CompanyID: creds.CompanyID,
			Verified:  false, // needs company admin approval
		}
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// POST /login
func Login(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", creds.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Enforce verification for normal users
	if user.Role == "user" && !user.Verified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Your account is pending approval from the company admin"})
		return
	}

	// Create JWT
	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserId:    user.ID,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"name":    user.Username,
			"email":   user.Email,
			"role":    user.Role,
			"company": user.CompanyID,
		},
	})
}

// PUT /users/:id/verify
func VerifyUser(c *gin.Context) {
	// Get user making request
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	// Only company admins (or super_admins) can verify
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can verify users"})
		return
	}

	// Parse target user
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Ensure target user is in same company (except super_admin)
	if role != "super_admin" && user.CompanyID != companyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only verify users in your company"})
		return
	}

	// Mark as verified
	user.Verified = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User verified successfully", "user": user})
}

// DELETE /users/:id/reject
func RejectUser(c *gin.Context) {
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	// Only admins or super_admins can reject users
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can reject users"})
		return
	}

	id := c.Param("id")
	var user models.User

	// Find user
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent rejecting users from other companies (unless super_admin)
	if role != "super_admin" && user.CompanyID != companyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to reject this user"})
		return
	}

	// Delete the user
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not reject user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User rejected successfully"})
}

// GET /users/pending
func GetPendingUsers(c *gin.Context) {
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	// Only admins or super_admins can view
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view pending users"})
		return
	}

	var users []models.User

	query := database.DB.Where("verified = ?", false)

	// Restrict to same company if not super_admin
	if role != "super_admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch pending users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_users": users})
}
