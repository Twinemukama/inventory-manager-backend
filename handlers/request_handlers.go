package handlers

import (
	"net/http"
	"strconv"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/models"
	"github.com/gin-gonic/gin"
)

func FetchPendingRequests(c *gin.Context) {
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	var pendingRequests []models.PendingRequest
	query := database.DB.Where("status = ?", "pending")
	// non-super admins should only see requests for their company
	if role != "super_admin" {
		query = query.Where("target_id = ?", companyID)
	}

	if err := query.Find(&pendingRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_requests": pendingRequests})
}

func RespondToRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// load the request
	var req models.PendingRequest
	if err := database.DB.First(&req, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	// authorization: only admins (of the target company) or super_admin can respond
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)
	if role != "super_admin" {
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can respond to requests"})
			return
		}
		if req.TargetID != companyID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to respond to this request"})
			return
		}
	}

	// process request: if accepted, mark user verified
	var processedUser *models.User
	if body.Status == "accepted" {
		var u models.User
		if err := database.DB.First(&u, req.UserID).Error; err == nil {
			u.Verified = true
			if err := database.DB.Save(&u).Error; err == nil {
				processedUser = &u
			}
		}
	}

	// delete the pending request after processing
	if err := database.DB.Delete(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pending request"})
		return
	}

	resp := gin.H{"message": "Request processed", "status": body.Status}
	if processedUser != nil {
		resp["user"] = processedUser
	}
	c.JSON(http.StatusOK, resp)
}
