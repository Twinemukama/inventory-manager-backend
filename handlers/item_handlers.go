package handlers

import (
	"net/http"
	"strconv"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/models"

	"github.com/gin-gonic/gin"
)

// POST /items
func CreateItem(c *gin.Context) {
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := c.MustGet("companyId").(uint)
	role := c.MustGet("role").(string)
	userID := c.MustGet("userId").(uint)

	item.UserID = userID

	// Only super admins might specify a company in the payload
	if role == "super_admin" && item.CompanyID != 0 {
		// use company provided in payload
	} else {
		// force user/admin item to their own company
		item.CompanyID = companyID
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GET /items
func ListItems(c *gin.Context) {
	// Parse query params
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	var items []models.Item
	var total int64

	query := database.DB.Model(&models.Item{}).Preload("User").Preload("Company")

	if role != "super_admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := query.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// GET /items/:id
func GetItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item

	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	query := database.DB.Where("id = ?", id)

	if role != "super_admin" {
		query = query.Preload("User").Preload("Company").Where("company_id = ?", companyID)
	}

	if err := query.First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// PUT /items/:id
func UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item

	userID := c.MustGet("userId").(uint)
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	if err := database.DB.First(&item, "id = ? AND company_id = ?", id, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if role != "admin" && role != "super_admin" && item.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this item"})
		return
	}
	var input models.Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update allowed fields
	item.Name = input.Name
	item.Description = input.Description
	item.Price = input.Price
	item.Quantity = input.Quantity
	item.CategoryID = input.CategoryID

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DELETE /items/:id
func DeleteItem(c *gin.Context) {
	id := c.Param("id")
	var item models.Item

	userID := c.MustGet("userId").(uint)
	role := c.MustGet("role").(string)
	companyID := c.MustGet("companyId").(uint)

	if err := database.DB.First(&item, "id = ? AND company_id = ?", id, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if role != "admin" && role != "super_admin" && item.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this item"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}
