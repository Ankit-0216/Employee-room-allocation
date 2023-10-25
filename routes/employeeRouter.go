package routes

import (
	controller "golang-employee-room-allocation/controllers"

	"github.com/gin-gonic/gin"
)

func EmployeeRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/employees", controller.GetEmployees())
	incomingRoutes.GET("/employees/:employee_id", controller.GetEmployee())
	incomingRoutes.POST("/employees", controller.CreateEmployee())
	incomingRoutes.POST("/employees/upload-csv", controller.CreateEmployeesFromCSV())
}