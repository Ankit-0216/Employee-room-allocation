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

	"encoding/csv"
	"io"
	// "golang.org/x/net/context"
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

// CreateEmployeesFromCSV creates employees from a CSV file
func CreateEmployeesFromCSV() gin.HandlerFunc {
    return func(c *gin.Context) {
        file, _, err := c.Request.FormFile("file") // Retrieve the uploaded file
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
            return
        }
        defer file.Close()

        // Parse the CSV file
        reader := csv.NewReader(file)
        var employees []interface{} // Use interface{} for data conversion
        for {
            record, err := reader.Read()
            if err == io.EOF {
                break
            }
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading the CSV file"})
                return
            }

            if len(record) != 2 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format, expected two columns"})
                return
            }

            employee := models.Employee{
                Employee_name: record[0],
                Nte_id:       record[1],
            }

            employee.Created_at = time.Now()
            employee.Updated_at = time.Now()
            employee.ID = primitive.NewObjectID()
            employee.Employee_id = employee.ID.Hex()
            employees = append(employees, employee)
        }

        // Insert employees into the database
        _, err = employeeCollection.InsertMany(context.Background(), employees) // Use context.Background()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting employees into the database"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": fmt.Sprintf("Successfully created %d employees", len(employees)),
        })
    }
}
