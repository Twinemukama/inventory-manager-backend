package main

import (
	"log"
	"net/http"

	"github.com/Twinemukama/go-inventory-manager/database"
	"github.com/Twinemukama/go-inventory-manager/handlers"
	"github.com/Twinemukama/go-inventory-manager/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	r := gin.Default()

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware())

	//sign up and login routes
	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)

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
