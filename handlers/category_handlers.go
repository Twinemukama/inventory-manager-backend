package handlers

import (
	"net/http"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/models"
	"github.com/gin-gonic/gin"
)

// POST /categories
func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("userId").(uint)       // get user ID from JWT
	companyID := c.MustGet("companyId").(uint) // get company ID from JWT
	role := c.MustGet("role").(string)

	category.UserID = userID

	if role == "super_admin" {
		// Super admin can assign company via payload
		if category.CompanyID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "company_id required for super_admin"})
			return
		}
	} else {
		// Force company_id from user context
		category.CompanyID = companyID
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GET /categories
func GetCategories(c *gin.Context) {
	var categories []models.Category

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	query := database.DB.Preload("User").Preload("Company")

	if role != "super_admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GET /categories/:id
func GetCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	query := database.DB.Where("id = ?", id)

	if role != "super_admin" {
		query = query.Preload("User").Preload("Company").Where("company_id = ?", companyID)
	}

	if err := query.First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// PUT /categories/:id
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)
	userID := c.MustGet("userId").(uint)

	query := database.DB.Where("id = ? AND user_id = ?", id, userID)
	if role != "super_admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var updated models.Category
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if role != "admin" && role != "super_admin" && category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot update this category"})
		return
	}

	category.Name = updated.Name

	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DELETE /categories/:id
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)
	userID := c.MustGet("userId").(uint)

	query := database.DB.Where("id = ? AND user_id = ?", id, userID)
	if role != "super_admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if role != "admin" && role != "super_admin" && category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete this category"})
		return
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}
