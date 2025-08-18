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
	userID := c.MustGet("userId").(uint) // get user ID from JWT
	category.UserID = userID

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GET /categories
func GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GET /categories/:id
func GetCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// PUT /categories/:id
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var updated models.Category
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = updated.Name
	category.UserID = updated.UserID

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

	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}
