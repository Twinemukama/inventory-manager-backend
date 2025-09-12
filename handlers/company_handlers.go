package handlers

import (
	"net/http"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/models"
	"github.com/gin-gonic/gin"
)

func GetCompanies(c *gin.Context) {
	var companies []models.Company

	if err := database.DB.Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, companies)
}
