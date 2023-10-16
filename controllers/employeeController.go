package controller

import (
	"context"
	"fmt"
	"golang-employee-room-allocation/database"
	"golang-employee-room-allocation/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var employeeCollection *mongo.Collection = database.OpenCollection(database.Client, "employee")

func GetEmployees() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := employeeCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the employees"})
		}
		var allemployees []bson.M
		if err = result.All(ctx, &allemployees); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allemployees)
	}
}

func GetEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		employeeId := c.Param("employee_id")
		var employee models.Employee

		err := employeeCollection.FindOne(ctx, bson.M{"employee_id": employeeId}).Decode(&employee)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching employee data"})
		}
		c.JSON(http.StatusOK, employee)
	}
}

func CreateEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		var employee models.Employee
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if err := c.BindJSON(&employee); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validationErr := validate.Struct(employee)
		// if validationErr != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		// 	return
		// }

		employee.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		employee.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		employee.ID = primitive.NewObjectID()
		employee.Employee_id = employee.ID.Hex()

		result, insertErr := employeeCollection.InsertOne(ctx, employee)
		if insertErr != nil {
			msg := fmt.Sprintf("Employee was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
		defer cancel()
	}
}

func UpdateEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var employee models.Employee

		if err := c.BindJSON(&employee); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		employeeId := c.Param("employee_id")
		filter := bson.M{"manuu_id": employeeId}

		var updateObj primitive.D

			if employee.Employee_name != "" {
				updateObj = append(updateObj, bson.E{"name", employee.Employee_name})
			}
			if employee.Nte_id != "" {
				updateObj = append(updateObj, bson.E{"nteId", employee.Nte_id})
			}

			employee.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{"updated_at", employee.Updated_at})

			upsert := true

			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := employeeCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{"$set", updateObj},
				},
				&opt,
			)

			if err != nil {
				msg := "Employee data update failed"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}

			defer cancel()
			c.JSON(http.StatusOK, result)
		
	}
}