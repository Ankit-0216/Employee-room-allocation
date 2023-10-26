package routes

import (
	controller "golang-employee-room-allocation/controllers"

	"github.com/gin-gonic/gin"
)

func RoomRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/rooms", controller.GetRooms())
	incomingRoutes.POST("/rooms", controller.CreateRoom())
	incomingRoutes.POST("/rooms/upload-csv", controller.CreateRoomsFromCSV())
	incomingRoutes.POST("/rooms/assign", controller.AssignEmployeeToRoom())
}