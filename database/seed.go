package database

import (
	"log"
	"os"

	"github.com/Twinemukama/go-inventory-manager/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedSuperAdmin() {
	var existing models.User
	if err := DB.Where("role = ?", "super_admin").First(&existing).Error; err == nil {
		return
	}

	company := models.Company{Name: "Aratech Dev"}
	if err := DB.FirstOrCreate(&company, models.Company{Name: "Aratech Dev"}).Error; err != nil {
		log.Println("Failed to create default company:", err)
		return
	}

	// Load super admin password from environment variable
	superAdminPassword := os.Getenv("SUPER_ADMIN_PASSWORD")
	if superAdminPassword == "" {
		log.Println("SUPER_ADMIN_PASSWORD environment variable is not set")
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(superAdminPassword), bcrypt.DefaultCost)
	superAdmin := models.User{
		Username:  "Twinemukama",
		Email:     "twinemukamai@gmail.com",
		Password:  string(hashedPassword),
		Role:      "super_admin",
		CompanyID: company.ID,
	}

	if err := DB.Create(&superAdmin).Error; err != nil {
		log.Println("Failed to create super admin:", err)
		return
	}

	log.Println("Super admin created successfully!")
}
