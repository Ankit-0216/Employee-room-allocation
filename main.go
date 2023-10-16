package main

import (
	"os"

	"golang-employee-room-allocation/database"

	middleware "golang-employee-room-allocation/middleware"
	routes "golang-employee-room-allocation/routes"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	routes.EmployeeRoutes(router)
	// routes.RoomRoutes(router)

	router.Run(":" + port)
}