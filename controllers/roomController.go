package controller

import (
	"context"
	"encoding/csv"
	"fmt"
	"golang-employee-room-allocation/database"
	"golang-employee-room-allocation/models"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var roomCollection *mongo.Collection = database.OpenCollection(database.Client, "room")

func GetRooms() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        result, err := roomCollection.Find(context.TODO(), bson.M{})
        defer cancel()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing the rooms"})
            return
        }
        var allRooms []bson.M
        if err = result.All(ctx, &allRooms); err != nil {
            log.Fatal(err)
        }
        c.JSON(http.StatusOK, allRooms)
    }
}

func CreateRoom() gin.HandlerFunc {
    return func(c *gin.Context) {
        var room models.Room
        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

        if err := c.BindJSON(&room); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Set the occupiedBy field to an empty slice initially
        room.OccupiedBy = []string{}

        room.ID = primitive.NewObjectID()
        result, insertErr := roomCollection.InsertOne(ctx, room)
        if insertErr != nil {
            msg := fmt.Sprintf("Room was not created")
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }
        defer cancel()
        c.JSON(http.StatusOK, result)
    }
}


// CreateRoomsFromCSV creates rooms from a CSV file
func CreateRoomsFromCSV() gin.HandlerFunc {
    return func(c *gin.Context) {
        file, _, err := c.Request.FormFile("file") // Retrieve the uploaded file
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
            return
        }
        defer file.Close()

        // Parse the CSV file
        reader := csv.NewReader(file)
        var rooms []interface{} // Use interface{} for data conversion
        for {
            record, err := reader.Read()
            if err == io.EOF {
                break
            }
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading the CSV file"})
                return
            }

            if len(record) != 3 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format, expected three columns"})
                return
            }

            room := models.Room{
                RoomNo:     record[0],
                RoomType:   record[1],
                OccupiedBy: []string{}, // Initialize as an empty slice
            }

            // Check if the room number is already occupied
            existingRoom := models.Room{}
            existingErr := roomCollection.FindOne(context.Background(), bson.M{"roomNo": room.RoomNo}).Decode(&existingRoom)
            if existingErr == nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Room is already occupied"})
                return
            }

            rooms = append(rooms, room)
        }

        // Insert rooms into the database
        _, err = roomCollection.InsertMany(context.Background(), rooms) // Use context.Background()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting rooms into the database"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": fmt.Sprintf("Successfully created %d rooms", len(rooms)),
        })
    }
}

func AssignEmployeeToRoom() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, _ = context.WithTimeout(context.Background(), 100*time.Second)

        // Fetch all available rooms
        allRooms := []models.Room{}
        roomCursor, err := roomCollection.Find(ctx, bson.M{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room data"})
            return
        }
        if err := roomCursor.All(ctx, &allRooms); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing room data"})
            return
        }

        // Fetch all available employees
        allEmployees := []models.Employee{}
        employeeCursor, err := employeeCollection.Find(ctx, bson.M{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employee data"})
            return
        }
        if err := employeeCursor.All(ctx, &allEmployees); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing employee data"})
            return
        }

        // Create a map to store rooms by type (single or double)
        roomsByType := make(map[string][]models.Room)
        for _, room := range allRooms {
            roomsByType[room.RoomType] = append(roomsByType[room.RoomType], room)
        }

        // Create a map to store employees by type (single or double)
        employeesByType := make(map[string][]models.Employee)
        for _, employee := range allEmployees {
            employeesByType[employee.Employee_id] = append(employeesByType[employee.Employee_id], employee)
        }

        // Iterate through room types and randomly assign employees
        updatedRooms := []models.Room{}
        unallocatedEmployees := []models.Employee{}

        for roomType, rooms := range roomsByType {
            employees, exists := employeesByType[roomType]

            if !exists || len(rooms) == 0 {
                continue
            }

            for i, room := range rooms {
                if i < len(employees) {
                    updatedRoom := room
                    updatedRoom.OccupiedBy = []string{employees[i].Employee_id}
                    updatedRooms = append(updatedRooms, updatedRoom)
                } else {
                    unallocatedEmployees = append(unallocatedEmployees, employees...)
                    break
                }
            }
        }

        // Update the rooms with assigned employees
        for _, updatedRoom := range updatedRooms {
            _, updateErr := roomCollection.UpdateOne(
                ctx,
                bson.M{"_id": updatedRoom.ID},
                bson.M{"$set": bson.M{"OccupiedBy": updatedRoom.OccupiedBy}},
            )
            if updateErr != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Error assigning employees to rooms"})
                return
            }
        }

        // Respond with the updated room data and unallocated employee IDs
        c.JSON(http.StatusOK, gin.H{
            "message":           "Employees assigned to rooms successfully",
            "updatedRoomData":   updatedRooms,
            "unallocatedEmployees": unallocatedEmployees,
        })
    }
}
