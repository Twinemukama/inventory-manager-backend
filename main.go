package main

import (
	"log"
	"net/http"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/handlers"
	"github.com/Twinemukama/go-inventory-manager/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	database.SeedSuperAdmin()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://inventory-manager-frontend-u13j.onrender.com",
			"http://localhost:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware())

	//sign up and login routes
	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)
	r.GET("/companies", handlers.GetCompanies)

	//verify user route - only admin and super admin can verify users
	auth.PUT("/users/:id/verify", handlers.VerifyUser)
	auth.GET("/users/pending", handlers.GetPendingUsers)
	auth.DELETE("/users/:id/reject", handlers.RejectUser)

	//Item routes
	auth.POST("/items", handlers.CreateItem)
	auth.GET("/items", handlers.ListItems)
	auth.GET("/items/:id", handlers.GetItem)
	auth.PUT("/items/:id", handlers.UpdateItem)
	auth.DELETE("/items/:id", handlers.DeleteItem)

	//Category routes
	auth.POST("/categories", handlers.CreateCategory)
	auth.GET("/categories", handlers.GetCategories)
	auth.GET("/categories/:id", handlers.GetCategory)
	auth.PUT("/categories/:id", handlers.UpdateCategory)
	auth.DELETE("/categories/:id", handlers.DeleteCategory)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
