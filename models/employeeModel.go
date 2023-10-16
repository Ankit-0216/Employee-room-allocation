package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	ID               primitive.ObjectID   `bson:"_id"`
	Employee_name    string               `json:"name" validate:"required,min=2,max=100"`
	Nte_id           string               `json:"nteId" validate:"required"`
	Created_at       time.Time            `json:"created_at"`
	Updated_at       time.Time            `json:"updated_at"`
	Employee_id      string               `json:"employee_id"`
}