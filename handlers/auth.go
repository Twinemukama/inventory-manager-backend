package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
	// Read raw body so we can handle alternate field names for company id
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var creds Credentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		// still try to continue; we'll return a bind error later if necessary
	}

	// also inspect raw map for alternate company id keys (companyID, company_id, company)
	var bodyMap map[string]interface{}
	_ = json.Unmarshal(raw, &bodyMap)
	if creds.CompanyID == 0 && bodyMap != nil {
		altKeys := []string{"companyId", "companyID", "company_id", "company"}
		for _, k := range altKeys {
			if v, ok := bodyMap[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					creds.CompanyID = uint(val)
				case string:
					if id, err := strconv.ParseUint(val, 10, 64); err == nil {
						creds.CompanyID = uint(id)
					}
				case int:
					creds.CompanyID = uint(val)
				}
				break
			}
		}

		// If the client sent a nested company object like { "company": { "id": 3 } }
		if creds.CompanyID == 0 {
			if compRaw, ok := bodyMap["company"]; ok && compRaw != nil {
				if compMap, ok := compRaw.(map[string]interface{}); ok {
					// look for id-like keys inside the company object
					for _, kid := range []string{"id", "companyId", "company_id"} {
						if v2, ok2 := compMap[kid]; ok2 && v2 != nil {
							switch val := v2.(type) {
							case float64:
								creds.CompanyID = uint(val)
							case string:
								if id, err := strconv.ParseUint(val, 10, 64); err == nil {
									creds.CompanyID = uint(id)
								}
							case int:
								creds.CompanyID = uint(val)
							}
							break
						}
					}
				}
			}
		}
	}

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

	if creds.CompanyID != 0 {
		var company models.Company
		if err := database.DB.First(&company, creds.CompanyID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Selected company does not exist"})
			return
		}

		user = models.User{
			Username:  creds.Username,
			Email:     creds.Email,
			Password:  string(hashedPassword),
			Role:      "user",
			CompanyID: company.ID,
			Verified:  false,
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
			return
		}

		pending := models.PendingRequest{
			UserID:      user.ID,
			UserName:    user.Username,
			UserEmail:   user.Email,
			Type:        "user_signup",
			TargetID:    company.ID,
			TargetName:  company.Name,
			Status:      "pending",
			Note:        "New user signup request for company approval",
			RequestedAt: time.Now(),
		}
		if err := database.DB.Create(&pending).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create pending request"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Signup successful. Waiting for company admin approval.",
			"status":  "pending_approval",
			"user":    user,
		})
		return
	}

	companyName := creds.CompanyName
	if companyName == "" {
		companyName = creds.Username + "'s Company"
	}

	// enforce uniqueness for new company creation at application level
	var existingCompany models.Company
	if err := database.DB.Where("name = ?", companyName).First(&existingCompany).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Company name already exists"})
		return
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

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Company created successfully. You are now the admin.",
		"status":  "approved",
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
